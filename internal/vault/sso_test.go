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

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/go-daemonkit/secrets"
)

func TestSSOPersistenceRecoveryAndNoPasswordFallback(t *testing.T) {
	data, err := os.ReadFile("../../testdata/apikey.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Email, Password string
		Token           map[string]any
		Sync            json.RawMessage
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	fixture.Token["refresh_token"] = strings.Repeat("sso-refresh.", 512)
	rotated := strings.Repeat("sso-rotated.", 1024)
	var snapshot struct{ Profile json.RawMessage }
	_ = json.Unmarshal(fixture.Sync, &snapshot)
	var rejected, failSync atomic.Bool
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/identity/connect/token":
			_ = r.ParseForm()
			if r.Form.Get("client_id") != "cli" || r.Form.Has("password") {
				t.Error("SSO client changed or password sent to server")
			}
			if rejected.Load() {
				w.WriteHeader(400)
				return
			}
			switch r.Form.Get("grant_type") {
			case "authorization_code":
				_ = json.NewEncoder(w).Encode(fixture.Token)
			case "refresh_token":
				_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "new-access", "refresh_token": rotated})
			default:
				t.Error("SSO fell back to another login method")
				w.WriteHeader(400)
			}
		case "/api/accounts/profile":
			_, _ = w.Write(snapshot.Profile)
		case "/api/sync":
			if failSync.Load() {
				w.WriteHeader(503)
				return
			}
			_, _ = w.Write(fixture.Sync)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	s, credentials := persistentService(t)
	defer s.Shutdown(context.Background())
	request := func() LoginSSORequest {
		return LoginSSORequest{LoginRequest: LoginRequest{Profile: "default", Server: server.URL, Password: []byte(fixture.Password)},
			SSOCode: bitwarden.SSOCode{Code: []byte("sso-code"), Verifier: []byte(strings.Repeat("v", 43)), RedirectURI: "http://localhost:8065"}}
	}
	req := request()
	if _, err := s.LoginSSO(t.Context(), req); err != nil {
		t.Fatal(err)
	}
	for _, secret := range [][]byte{req.Password, req.Code, req.Verifier} {
		if !bytes.Equal(secret, make([]byte, len(secret))) {
			t.Fatal("SSO RPC input secret survived login")
		}
	}
	saved, err := s.load(t.Context(), "default")
	if err != nil || saved.Version != 4 || saved.AuthMethod != "sso" || saved.Cache.Email != fixture.Email || saved.Cache.OAuthClientID != "cli" {
		t.Fatal("SSO identity or login method not persisted", err)
	}
	encoded, _ := json.Marshal(saved)
	for _, secret := range []string{fixture.Password, "sso-refresh", "test-access", "sso-code", string(request().Verifier)} {
		if bytes.Contains(encoded, []byte(secret)) || bytes.Contains(encoded, []byte(base64.StdEncoding.EncodeToString([]byte(secret)))) {
			t.Fatal("secret persisted in cache")
		}
	}
	oldSession, oldCredential := s.profiles["default"].session, saved.Credential
	for _, kind := range []string{"wrong password", "wrong account", "keychain failure"} {
		req := request()
		switch kind {
		case "wrong password":
			req.Password = []byte("wrong password")
		case "wrong account":
			req.Email = "other@example.invalid"
		case "keychain failure":
			credentials.setErr = secrets.ErrAccess
		}
		if _, err := s.LoginSSO(t.Context(), req); err == nil {
			t.Fatal("accepted", kind)
		}
		credentials.setErr = nil
		current, err := s.load(t.Context(), "default")
		if err != nil || current.Credential != oldCredential || s.profiles["default"].session != oldSession {
			t.Fatal("failed login replaced saved session", kind, err)
		}
	}
	if err := s.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}
	restored := &Service{Store: s.Store, Secrets: credentials}
	defer restored.Shutdown(context.Background())
	before := requests.Load()
	credentials.getErr = secrets.ErrAccess
	if _, err := restored.Unlock(t.Context(), UnlockRequest{Profile: "default", Password: []byte(fixture.Password)}); err != nil {
		t.Fatal("offline unlock accessed credential store", err)
	}
	if requests.Load() != before {
		t.Fatal("offline unlock made network request")
	}
	credentials.getErr = nil
	failSync.Store(true)
	if _, err := restored.Sync(t.Context(), ProfileRequest{Profile: "default"}); err == nil {
		t.Fatal("sync failure ignored")
	}
	saved, _ = restored.load(t.Context(), "default")
	auth, err := restored.loadCredential(t.Context(), saved)
	if err != nil || string(auth.Token) != rotated || auth.ClientID != "" || len(auth.ClientSecret) != 0 {
		t.Fatal("token rotation was lost after sync failure", err)
	}
	auth.clear()
	failSync.Store(false)
	if _, err := restored.Sync(t.Context(), ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Lock(t.Context(), ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Unlock(t.Context(), UnlockRequest{Profile: "default", Password: []byte(fixture.Password)}); err != nil {
		t.Fatal(err)
	}
	rejected.Store(true)
	_, err = restored.Sync(t.Context(), ProfileRequest{Profile: "default"})
	var needed *AuthRequired
	if !errors.As(err, &needed) || needed.Method != "sso" || needed.Action != "login" || !strings.Contains(err.Error(), "login --sso") {
		t.Fatal("expired SSO did not require explicit SSO login", err)
	}
	before = requests.Load()
	if _, err := restored.Login(t.Context(), request().LoginRequest); !errors.As(err, &needed) || needed.Method != "sso" {
		t.Fatal("SSO fell back to password login", err)
	}
	if _, err := restored.BeginAuth(t.Context(), AuthRequest{Profile: "default", Action: "login", Epoch: needed.Epoch, Version: needed.Version}); !errors.As(err, &needed) {
		t.Fatal("SSO incorrectly reserved a password prompt", err)
	}
	if before != requests.Load() {
		t.Fatal("fallback reached server")
	}
	rejected.Store(false)
	if _, err := restored.LoginSSO(t.Context(), request()); err != nil {
		t.Fatal("explicit SSO relogin failed", err)
	}
	if _, err := restored.Logout(t.Context(), ProfileRequest{Profile: "default"}); err != nil || len(credentials.values) != 0 {
		t.Fatal("logout retained SSO credentials", err)
	}
}
