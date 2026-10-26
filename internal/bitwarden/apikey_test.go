package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
	}{
		{"missing token", func(m map[string]any) { delete(m, "access_token") }},
		{"missing KDF", func(m map[string]any) { delete(m, "Kdf") }},
		{"argon2", func(m map[string]any) { m["Kdf"] = 1 }},
		{"excessive KDF", func(m map[string]any) { m["KdfIterations"] = 2000001 }},
		{"missing key", func(m map[string]any) { delete(m, "Key") }},
		{"invalid key", func(m map[string]any) { m["Key"] = "2.invalid" }},
		{"key connector", func(m map[string]any) { m["ApiUseKeyConnector"] = true }},
		{"no master password", func(m map[string]any) { m["UserDecryptionOptions"] = map[string]any{"HasMasterPassword": false} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, _ := json.Marshal(response)
			var body map[string]any
			_ = json.Unmarshal(data, &body)
			test.change(body)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(body) }))
			defer server.Close()
			c, _ := NewClient(server.URL, email, testDevice)
			if s, err := c.LoginAPIKey(t.Context(), []byte(password), key); err == nil || s != nil {
				t.Fatal("invalid response accepted")
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
