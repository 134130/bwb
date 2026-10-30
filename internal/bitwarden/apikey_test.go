package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func apiKeyFixture(t *testing.T) (APIKey, string, string, tokens, syncResponse) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/apikey.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Email, Password, ClientID, ClientSecret string
		Token                                   tokens
		Sync                                    syncResponse
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	return APIKey{ClientID: f.ClientID, ClientSecret: []byte(f.ClientSecret)}, f.Email, f.Password, f.Token, f.Sync
}

func TestAPIKeyDiscoversEmailFromAuthenticatedProfile(t *testing.T) {
	key, email, password, token, snapshot := apiKeyFixture(t)
	for _, tc := range []struct {
		name, id, email string
		status          int
		valid           bool
	}{
		{"normalized email", snapshot.Profile.ID, " " + strings.ToUpper(email) + " ", http.StatusOK, true},
		{"missing email", snapshot.Profile.ID, "", http.StatusOK, false},
		{"invalid email", snapshot.Profile.ID, "not-an-email", http.StatusOK, false},
		{"display name", snapshot.Profile.ID, "Name <" + email + ">", http.StatusOK, false},
		{"wrong account", testDevice, email, http.StatusOK, false},
		{"missing account", "", email, http.StatusOK, false},
		{"profile rejected", snapshot.Profile.ID, email, http.StatusUnauthorized, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var profileCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/identity/connect/token":
					_ = json.NewEncoder(w).Encode(token)
				case "/api/accounts/profile":
					profileCalls.Add(1)
					if r.Method != "GET" || r.Header.Get("Authorization") != "Bearer "+token.Access {
						t.Error("profile discovery was not authenticated")
					}
					w.WriteHeader(tc.status)
					_ = json.NewEncoder(w).Encode(map[string]string{"id": tc.id, "email": tc.email})
				case "/api/sync":
					_ = json.NewEncoder(w).Encode(snapshot)
				default:
					t.Error("unexpected endpoint", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "", testDevice)
			if err != nil {
				t.Fatal(err)
			}
			session, err := client.LoginAPIKey(t.Context(), []byte(password), key)
			if (err == nil) != tc.valid || profileCalls.Load() != 1 || client.email != "" {
				t.Fatal("unexpected discovery result or mutated client", err)
			}
			if !tc.valid {
				if session != nil {
					session.Close()
					t.Fatal("invalid profile created a session")
				}
				return
			}
			defer session.Close()
			got, err := session.SyncSnapshot(t.Context())
			if err != nil || session.Cache(got).Email != email {
				t.Fatal("discovered email did not bind the cache", err)
			}
		})
	}
}

func TestAPIKeyLoginRenewUnlockAndClose(t *testing.T) {
	key, email, password, token, snapshot := apiKeyFixture(t)
	var tokenCalls, syncCalls atomic.Int32
	var reject atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/identity/connect/token":
			tokenCalls.Add(1)
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("scope") != "api" ||
				r.Form.Get("client_id") != key.ClientID || r.Form.Get("client_secret") != string(key.ClientSecret) ||
				r.Form.Get("deviceType") != map[string]string{"windows": "23", "darwin": "24", "linux": "25"}[runtime.GOOS] ||
				r.Form.Has("password") || r.Form.Has("username") || r.Form.Has("refresh_token") || r.Form.Get("deviceIdentifier") != testDevice {
				t.Error("incorrect API key authentication request")
			}
			if reject.Load() {
				w.WriteHeader(400)
				return
			}
			_ = json.NewEncoder(w).Encode(token)
		case "/api/sync":
			n := syncCalls.Add(1)
			if r.Header.Get("Authorization") != "Bearer test-access" {
				t.Error("missing access token")
			}
			if n == 2 || reject.Load() {
				w.WriteHeader(401)
				return
			}
			_ = json.NewEncoder(w).Encode(snapshot)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c, _ := NewClient(server.URL, email, testDevice)
	s, err := c.LoginAPIKey(t.Context(), []byte(password), key)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	first, err := s.SyncSnapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	visible, err := s.decryptSnapshot(t.Context(), snapshot)
	if err != nil || len(visible.Items) != 2 {
		t.Fatal("reprompt item blocked snapshot", err)
	}
	restricted := visible.Items[0]
	if restricted.Name != "Reprompt item" || restricted.Data != nil || restricted.Fields != nil || restricted.Notes != nil {
		t.Fatal("reprompt item exposed fields")
	}
	if value, err := visible.Read(restricted.ID, "login/password"); !errors.Is(err, ErrReprompt) || value != "" {
		t.Fatal("vault read bypassed reprompt", err)
	}
	// Unreadable secret fields remain encrypted; a tampered display name still rejects the snapshot.
	var item map[string]any
	_ = json.Unmarshal(snapshot.Ciphers[0], &item)
	item["name"] = "2.invalid"
	broken := snapshot
	broken.Ciphers = append([]json.RawMessage(nil), snapshot.Ciphers...)
	broken.Ciphers[0], _ = json.Marshal(item)
	if _, _, err := s.validateSnapshot(t.Context(), broken); !errors.Is(err, ErrCipher) {
		t.Fatal("reprompt bypassed name authentication", err)
	}
	if s.refresh != nil || !s.HasCredentials() {
		t.Fatal("API key session requires a refresh token")
	}
	if _, err := s.SyncSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	if tokenCalls.Load() != 2 || syncCalls.Load() != 3 {
		t.Fatal("401 did not renew exactly once")
	}
	cache := s.Cache(first)
	oldKey, oldAccess, oldSecret := s.key, s.access, s.apiSecret
	s.Close()
	if oldKey.IsAlive() || oldAccess.IsAlive() || oldSecret.IsAlive() || s.HasCredentials() {
		t.Fatal("secrets survived Close")
	}
	if string(key.ClientSecret) != "PUBLIC-TEST-API-KEY" {
		t.Fatal("session mutated caller credential")
	}
	before := tokenCalls.Load() + syncCalls.Load()
	offline, _, err := Unlock(t.Context(), cache, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer offline.Close()
	if offline.HasCredentials() || before != tokenCalls.Load()+syncCalls.Load() {
		t.Fatal("unlock used network credentials")
	}
	wrong := key
	wrong.ClientID = "user." + testDevice
	if err := offline.SetAPIKey(wrong); err == nil {
		t.Fatal("accepted another account's key")
	}
	if err := offline.SetAPIKey(key); err != nil {
		t.Fatal(err)
	}
	if _, err := offline.SyncSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	reject.Store(true)
	before = tokenCalls.Load()
	if _, err := offline.SyncSnapshot(t.Context()); !errors.Is(err, ErrAuthentication) {
		t.Fatal(err)
	}
	if tokenCalls.Load() != before+1 {
		t.Fatal("unbounded API key retry")
	}
	offline.Close()
	if err := offline.SetAPIKey(key); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed session accepted key")
	}
}

func TestAPIKeyRejectsInvalidResponsesAndCredentials(t *testing.T) {
	key, email, password, response, snapshot := apiKeyFixture(t)
	for _, test := range []struct {
		name   string
		change func(map[string]any)
		want   string
	}{
		{"missing token", func(m map[string]any) { delete(m, "access_token") }, "incomplete API key login response"},
		{"missing KDF", func(m map[string]any) { delete(m, "Kdf") }, "missing Kdf"},
		{"argon2", func(m map[string]any) { m["Kdf"] = 1 }, "Argon2id requires"},
		{"excessive KDF", func(m map[string]any) { m["KdfIterations"] = 2000001 }, "PBKDF2 iterations"},
		{"missing key", func(m map[string]any) { delete(m, "Key") }, "missing the protected vault key"},
		{"invalid key", func(m map[string]any) { m["Key"] = "2.invalid" }, "decrypt protected vault key"},
		{"key connector", func(m map[string]any) { m["ApiUseKeyConnector"] = true }, "requires a master password"},
		{"no master password", func(m map[string]any) { m["UserDecryptionOptions"] = map[string]any{"HasMasterPassword": false} }, "requires a master password"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, _ := json.Marshal(response)
			var body map[string]any
			_ = json.Unmarshal(data, &body)
			test.change(body)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(body) }))
			defer server.Close()
			c, _ := NewClient(server.URL, email, testDevice)
			if s, err := c.LoginAPIKey(t.Context(), []byte(password), key); err == nil || s != nil || !strings.Contains(err.Error(), test.want) {
				t.Fatal("missing response diagnostic", err)
			}
		})
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/api/sync" {
			wrong := snapshot
			wrong.Profile.ID = testDevice
			_ = json.NewEncoder(w).Encode(wrong)
		} else {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	c, _ := NewClient(server.URL, email, testDevice)
	for _, invalid := range []APIKey{{}, {ClientID: strings.Replace(key.ClientID, "user.", "organization.", 1), ClientSecret: key.ClientSecret}, {ClientID: key.ClientID}, {ClientID: key.ClientID, ClientSecret: []byte(strings.Repeat("x", 4097))}} {
		if s, err := c.LoginAPIKey(t.Context(), []byte(password), invalid); err == nil || s != nil {
			t.Fatal("invalid API key accepted")
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid API key reached network")
	}
	if s, err := c.LoginAPIKey(t.Context(), []byte("wrong password"), key); err == nil || s != nil {
		t.Fatal("wrong password accepted")
	}
	s, err := c.LoginAPIKey(t.Context(), []byte(password), key)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.SyncSnapshot(t.Context()); err == nil {
		t.Fatal("API key account mismatch accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := s.Refresh(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
