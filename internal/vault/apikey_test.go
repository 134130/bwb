package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/go-daemonkit/secrets"
)

func apiService(t *testing.T) (*Service, *faultyCredentials, func() LoginAPIKeyRequest, *atomic.Bool, *atomic.Int32) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/apikey.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Email, Password, ClientID, ClientSecret string
		Token, Sync                             json.RawMessage
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var rejected atomic.Bool
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/identity/connect/token":
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "client_credentials" {
				t.Error("password login or refresh fallback")
			}
			if rejected.Load() || r.Form.Get("client_secret") != fixture.ClientSecret {
				w.WriteHeader(400)
				return
			}
			_, _ = w.Write(fixture.Token)
		case "/api/sync":
			_, _ = w.Write(fixture.Sync)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(server.Close)
	s, f := persistentService(t)
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })
	request := func() LoginAPIKeyRequest {
		return LoginAPIKeyRequest{LoginRequest: LoginRequest{Profile: "default", Server: server.URL, Email: fixture.Email, Password: []byte(fixture.Password)},
			APIKey: bitwarden.APIKey{ClientID: fixture.ClientID, ClientSecret: []byte(fixture.ClientSecret)}}
	}
	return s, f, request, &rejected, &requests
}

func TestAPIKeyPersistenceRecoveryAndNoPasswordFallback(t *testing.T) {
	s, f, request, rejected, requests := apiService(t)
	ctx := t.Context()
	req := request()
	if _, err := s.LoginAPIKey(ctx, req); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(req.Password, make([]byte, len(req.Password))) || !bytes.Equal(req.ClientSecret, make([]byte, len(req.ClientSecret))) {
		t.Fatal("RPC input secrets survived login")
	}
	saved, err := s.load(ctx, "default")
	if err != nil || saved.Version != 2 || saved.AuthMethod != "apikey" {
		t.Fatal("API key metadata missing", err)
	}
	auth, err := s.loadCredential(ctx, saved)
	if err != nil || auth.ClientID != request().ClientID || string(auth.ClientSecret) != string(request().ClientSecret) || len(auth.Token) != 0 {
		t.Fatal("API credential not persisted", err)
	}
	auth.clear()
	data, _ := json.Marshal(saved)
	for _, secret := range []string{string(request().ClientSecret), base64.StdEncoding.EncodeToString(request().ClientSecret), string(request().Password), "test-access"} {
		if bytes.Contains(data, []byte(secret)) {
			t.Fatal("secret in cache record")
		}
	}
	oldCredential := saved.Credential
	oldSession := s.profiles["default"].session
	for _, kind := range []string{"bad key", "bad password", "wrong account", "keychain"} {
		req := request()
		switch kind {
		case "bad key":
			req.ClientSecret = []byte("wrong")
		case "bad password":
			req.Password = []byte("wrong")
		case "wrong account":
			req.ClientID = "user.30000000-0000-4000-8000-000000000001"
		case "keychain":
			f.setErr = secrets.ErrAccess
		}
		if _, err := s.LoginAPIKey(ctx, req); err == nil {
			t.Fatal("accepted", kind)
		}
		f.setErr = nil
		current, err := s.load(ctx, "default")
		if err != nil || current.Credential != oldCredential || s.profiles["default"].session != oldSession {
			t.Fatal("failed login replaced previous session", kind, err)
		}
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	restored := &Service{Store: s.Store, Secrets: f}
	defer restored.Shutdown(context.Background())
	f.getErr = secrets.ErrAccess
	before := requests.Load()
	if _, err := restored.Unlock(ctx, UnlockRequest{Profile: "default", Password: request().Password}); err != nil {
		t.Fatal("offline unlock consulted keychain", err)
	}
	if requests.Load() != before {
		t.Fatal("unlock contacted server")
	}
	if _, err := restored.Sync(ctx, ProfileRequest{Profile: "default"}); !errors.Is(err, secrets.ErrAccess) {
		t.Fatal("keychain error hidden", err)
	}
	f.getErr = nil
	f.setErr = secrets.ErrAccess
	if _, err := restored.Sync(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal("API sync rewrote credentials", err)
	}
	f.setErr = nil
	if _, err := restored.Lock(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Unlock(ctx, UnlockRequest{Profile: "default", Password: request().Password}); err != nil {
		t.Fatal(err)
	}
	rejected.Store(true)
	_, err = restored.Sync(ctx, ProfileRequest{Profile: "default"})
	var needed *AuthRequired
	if !errors.As(err, &needed) || needed.Method != "apikey" || needed.Action != "login" || !strings.Contains(err.Error(), "--apikey") {
		t.Fatal("missing API reauthentication hint", err)
	}
	if _, err := restored.BeginAuth(ctx, AuthRequest{Profile: "default", Action: "login", Epoch: needed.Epoch, Version: needed.Version}); !errors.As(err, &needed) {
		t.Fatal("API key failure opened a password prompt", err)
	}
	if _, err := restored.Login(ctx, request().LoginRequest); !errors.As(err, &needed) {
		t.Fatal("password fallback allowed", err)
	}
	before = requests.Load()
	if _, err := restored.List(ctx, ListRequest{Profile: "default", Kind: "folder"}); err != nil {
		t.Fatal("revocation destroyed cached reads", err)
	}
	if requests.Load() != before {
		t.Fatal("cached list checked server credentials")
	}
	rejected.Store(false)
	if _, err := restored.LoginAPIKey(ctx, request()); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Logout(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if len(f.values) != 0 {
		t.Fatal("logout left API keys")
	}
	if _, err := restored.Unlock(ctx, UnlockRequest{Profile: "default", Password: request().Password}); !errors.As(err, &needed) {
		t.Fatal("logout retained unlock", err)
	}
}

func TestAPIKeyCredentialBindingAndVersion(t *testing.T) {
	s, f, request, _, _ := apiService(t)
	if _, err := s.LoginAPIKey(t.Context(), request()); err != nil {
		t.Fatal(err)
	}
	saved, _ := s.load(t.Context(), "default")
	original := bytes.Clone(f.values[saved.Credential])
	for _, field := range []string{"Server", "Email", "AccountID", "ClientID", "Token", "ClientSecret"} {
		var changed map[string]any
		_ = json.Unmarshal(original, &changed)
		changed[field] = "wrong"
		if field == "Token" || field == "ClientSecret" {
			changed[field] = base64.StdEncoding.EncodeToString([]byte("wrong"))
		}
		if field == "ClientSecret" {
			changed[field] = ""
		}
		f.values[saved.Credential], _ = json.Marshal(changed)
		if auth, err := s.loadCredential(t.Context(), saved); err == nil {
			auth.clear()
			t.Fatal("accepted corrupted credential", field)
		}
	}
	f.values[saved.Credential] = original
	if err := s.saveToken(t.Context(), saved, []byte("refresh")); err == nil || !bytes.Equal(original, f.values[saved.Credential]) {
		t.Fatal("refresh overwrote API key")
	}
	saved.Version = 1
	if err := s.put(t.Context(), "default", saved); err != nil {
		t.Fatal(err)
	}
	if _, err := s.load(t.Context(), "default"); err == nil {
		t.Fatal("v1 accepted API key metadata")
	}
}

func TestAPIKeySyncCanceledByLockAndLogout(t *testing.T) {
	for _, logout := range []bool{false, true} {
		t.Run(map[bool]string{false: "lock", true: "logout"}[logout], func(t *testing.T) {
			s, f, request, _, requests := apiService(t)
			ctx := context.Background()
			if _, err := s.LoginAPIKey(ctx, request()); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Lock(ctx, ProfileRequest{Profile: "default"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Unlock(ctx, UnlockRequest{Profile: "default", Password: request().Password}); err != nil {
				t.Fatal(err)
			}
			before := requests.Load()
			synctest.Test(t, func(t *testing.T) {
				// A channel created outside the bubble is not durably blocking to synctest.
				gate := s.profiles["default"].gate
				defer func() { s.profiles["default"].gate = gate }()
				s.profiles["default"].gate = make(chan struct{}, 1)
				release := make(chan struct{})
				f.beforeGet = func() { <-release }
				defer func() { f.beforeGet = nil }()
				synced := make(chan error, 1)
				go func() { _, err := s.Sync(ctx, ProfileRequest{Profile: "default"}); synced <- err }()
				synctest.Wait()
				cleared := make(chan error, 1)
				go func() { cleared <- s.clearProfile(ctx, "default", logout) }()
				synctest.Wait()
				select {
				case <-cleared:
					t.Fatal("lock did not wait for credential read")
				default:
				}
				close(release)
				synctest.Wait()
				if err := <-synced; !errors.Is(err, context.Canceled) {
					t.Fatal("sync survived lock", err)
				}
				if err := <-cleared; err != nil {
					t.Fatal(err)
				}
			})
			if requests.Load() != before || s.profiles["default"].session != nil {
				t.Fatal("late network request or publication")
			}
			if logout && len(f.values) != 0 {
				t.Fatal("logout left API key")
			}
		})
	}
}
