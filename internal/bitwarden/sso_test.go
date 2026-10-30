package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestVaultwardenSSODiscoveryChecksPlaceholder(t *testing.T) {
	for _, tc := range []struct {
		name, body    string
		status, count int
		fail          bool
	}{
		{"enabled", `{"token":"validation-token"}`, 200, 1, false},
		{"disabled", `{}`, 400, 0, false},
		{"outage", `{}`, 503, 0, true},
		{"invalid", `{}`, 200, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const identifier = "00000000-01DC-01DC-01DC-000000000000"
			var validations atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/organizations/domain/sso/verified" {
					_, _ = w.Write([]byte(`{"data":[{"organizationIdentifier":"` + identifier + `"}]}`))
					return
				}
				validations.Add(1)
				if r.URL.Path != "/identity/sso/prevalidate" || r.Method != "GET" || r.URL.Query().Get("domainHint") != identifier {
					t.Error("invalid SSO prevalidation")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "user@example.invalid", "00000000-0000-4000-8000-000000000001")
			if err != nil {
				t.Fatal(err)
			}
			identifiers, err := client.DiscoverSSO(t.Context())
			if (err != nil) != tc.fail || len(identifiers) != tc.count || validations.Load() != 1 {
				t.Fatal("placeholder was treated as an SSO policy", identifiers, err)
			}
		})
	}
}

func TestSSODiscovery(t *testing.T) {
	for _, tc := range []struct {
		name, body    string
		status, count int
		want          error
	}{
		{"match", `{"data":[{"organizationIdentifier":"work"}]}`, 200, 1, nil},
		{"multiple", `{"data":[{"organizationIdentifier":"one"},{"organizationIdentifier":"two"}]}`, 200, 2, nil},
		{"none", `{"data":[]}`, 200, 0, nil},
		{"unsupported", ``, 404, 0, ErrSSODiscoveryUnsupported},
		{"method unsupported", ``, 405, 0, ErrSSODiscoveryUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var input struct{ Email string }
				_ = json.NewDecoder(r.Body).Decode(&input)
				if r.Method != "POST" || r.URL.Path != "/api/organizations/domain/sso/verified" || input.Email != "user@example.invalid" || r.Header.Get("Authorization") != "" {
					t.Error("incorrect discovery request")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client, _ := NewClient(server.URL, " User@example.invalid ", testDevice)
			ids, err := client.DiscoverSSO(t.Context())
			if !errors.Is(err, tc.want) || len(ids) != tc.count {
				t.Fatal(ids, err)
			}
		})
	}
	for _, body := range []string{`{}`, `{"data":null}`, `{"data":[{}]}`, `{"data":[{"organizationIdentifier":"\n"}]}`, `bad-json`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
		client, _ := NewClient(server.URL, "user@example.invalid", testDevice)
		if _, err := client.DiscoverSSO(t.Context()); err == nil || errors.Is(err, ErrSSODiscoveryUnsupported) {
			t.Fatal("malformed discovery fell back", err)
		}
		server.Close()
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer server.Close()
	client, _ := NewClient(server.URL, "user@example.invalid", testDevice)
	if _, err := client.DiscoverSSO(t.Context()); err == nil || errors.Is(err, ErrSSODiscoveryUnsupported) {
		t.Fatal("outage treated as unsupported", err)
	}
}

func TestSSOExchangeBeforePasswordAndRetry(t *testing.T) {
	_, email, password, token, snapshot := apiKeyFixture(t)
	token.Refresh = "sso-refresh"
	var exchanges atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identity/connect/token" {
			exchanges.Add(1)
			_ = json.NewEncoder(w).Encode(token)
		} else {
			_ = json.NewEncoder(w).Encode(snapshot.Profile)
		}
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, email, testDevice)
	code := testSSOCode()
	code.Identifier = "work"
	pending, err := client.ExchangeSSO(t.Context(), code)
	if err != nil {
		t.Fatal(err)
	}
	defer pending.Close()
	access, refresh := pending.access, pending.refresh
	if pending.active() || pending.Email() != email {
		t.Fatal("exchange unlocked vault")
	}
	if _, err := pending.UnlockSSO(t.Context(), []byte("wrong")); !errors.Is(err, ErrCipher) {
		t.Fatal(err)
	}
	session, err := pending.UnlockSSO(t.Context(), []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if exchanges.Load() != 1 || session.ssoIdentifier != "work" || !session.active() {
		t.Fatal("password retry repeated browser authentication")
	}
	pending.Close()
	if access.IsAlive() || refresh.IsAlive() || !session.active() {
		t.Fatal("pending credentials not independently destroyed")
	}
	if _, err := pending.UnlockSSO(t.Context(), []byte(password)); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed login accepted", err)
	}
}

func testSSOCode() SSOCode {
	return SSOCode{Code: []byte("test-code"), Verifier: []byte(strings.Repeat("v", 43)), RedirectURI: "http://localhost:8065"}
}

func TestSSOLoginRefreshAndOfflineUnlock(t *testing.T) {
	t.Run("PBKDF2", func(t *testing.T) { testSSOLoginRefreshAndOfflineUnlock(t, false) })
	t.Run("Argon2id", func(t *testing.T) { testSSOLoginRefreshAndOfflineUnlock(t, true) })
}

func testSSOLoginRefreshAndOfflineUnlock(t *testing.T, argon bool) {
	_, email, password, token, snapshot := apiKeyFixture(t)
	if argon {
		old, _, err := derive(t.Context(), email, []byte(password), token.kdfConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer clear(old)
		key, err := decrypt(token.Key, old)
		if err != nil {
			t.Fatal(err)
		}
		defer clear(key)
		kind := 1
		token.KDF, token.Iterations, token.Memory, token.Parallelism = &kind, 3, 64, 4
		stretched, _, err := derive(t.Context(), email, []byte(password), token.kdfConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer clear(stretched)
		padded := append(bytes.Clone(key), bytes.Repeat([]byte{16}, 16)...)
		defer clear(padded)
		token.Key = encryptPadded(t, padded, stretched)
		snapshot.Profile.Key = token.Key
	}
	token.Refresh = "sso-refresh"
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/identity/connect/token":
			_ = r.ParseForm()
			if r.Form.Get("client_id") != "cli" || r.Form.Has("password") || r.Form.Has("username") || r.Form.Has("client_secret") {
				t.Error("wrong SSO client or password sent to server")
			}
			switch r.Form.Get("grant_type") {
			case "authorization_code":
				if r.Form.Get("code") != "test-code" || r.Form.Get("code_verifier") != string(testSSOCode().Verifier) ||
					r.Form.Get("redirect_uri") != testSSOCode().RedirectURI || r.Form.Get("scope") != "api offline_access" ||
					r.Form.Get("deviceIdentifier") != testDevice || r.Form.Get("deviceType") != cliDeviceType() {
					t.Error("invalid SSO code exchange")
				}
				_ = json.NewEncoder(w).Encode(token)
			case "refresh_token":
				if r.Form.Get("refresh_token") != token.Refresh {
					t.Error("wrong refresh token")
				}
				_ = json.NewEncoder(w).Encode(tokens{Access: "new-access", Refresh: "rotated"})
			default:
				t.Error("unexpected grant")
				w.WriteHeader(400)
			}
		case "/api/accounts/profile":
			if r.Header.Get("Authorization") != "Bearer "+token.Access {
				t.Error("unauthenticated profile lookup")
			}
			_ = json.NewEncoder(w).Encode(snapshot.Profile)
		case "/api/sync":
			_ = json.NewEncoder(w).Encode(snapshot)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "", testDevice)
	code := testSSOCode()
	code.Identifier = "work"
	session, err := client.LoginSSO(t.Context(), []byte(password), code)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if client.email != "" || session.client.email != email {
		t.Fatal("SSO did not resolve email on a separate client")
	}
	got, err := session.SyncSnapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	cache := session.Cache(got)
	if cache.OAuthClientID != "cli" || cache.SSOIdentifier != "work" || cache.KDF != *token.KDF || cache.Memory != token.Memory || cache.Parallelism != token.Parallelism {
		t.Fatal("SSO/KDF metadata missing from cache")
	}
	key, access, refresh := session.key, session.access, session.refresh
	session.Close()
	if key.IsAlive() || access.IsAlive() || refresh.IsAlive() {
		t.Fatal("secrets survived Close")
	}
	before := requests.Load()
	restored, _, err := Unlock(t.Context(), cache, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	if restored.ssoIdentifier != "work" {
		t.Fatal("lost SSO identifier")
	}
	if err := restored.VerifyPassword(t.Context(), []byte(password)); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != before || restored.HasCredentials() {
		t.Fatal("unlock accessed network credentials")
	}
	if err := restored.SetRefreshToken([]byte(token.Refresh)); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.SyncSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	if string(restored.RefreshToken()) != "rotated" {
		t.Fatal("refresh did not rotate")
	}
	cache.OAuthClientID = "unexpected-client"
	if _, _, err := Unlock(t.Context(), cache, []byte(password)); !errors.Is(err, ErrUnsupported) {
		t.Fatal("accepted unknown OAuth client", err)
	}
}

func TestSSORejectsInvalidAccountAndResponse(t *testing.T) {
	_, email, password, token, snapshot := apiKeyFixture(t)
	token.Refresh = "sso-refresh"
	for _, tc := range []struct {
		name, want string
		change     func(map[string]any)
	}{
		{"missing access", "incomplete", func(m map[string]any) { delete(m, "access_token") }},
		{"missing refresh", "incomplete", func(m map[string]any) { delete(m, "refresh_token") }},
		{"missing KDF", "missing Kdf", func(m map[string]any) { delete(m, "Kdf") }},
		{"invalid argon2", "Argon2id requires", func(m map[string]any) { m["Kdf"] = 1 }},
		{"excessive KDF", "iterations", func(m map[string]any) { m["KdfIterations"] = 2000001 }},
		{"missing key", "protected vault key", func(m map[string]any) { delete(m, "Key") }},
		{"bad key", "decrypt protected vault key", func(m map[string]any) { m["Key"] = "2.invalid" }},
		{"key connector", "requires a master password", func(m map[string]any) { m["KeyConnectorUrl"] = "https://example.invalid" }},
		{"passwordless", "requires a master password", func(m map[string]any) { m["UserDecryptionOptions"] = map[string]any{"HasMasterPassword": false} }},
		{"wrong account", "email differs", func(map[string]any) {}},
		{"wrong password", "decrypt protected vault key", func(map[string]any) {}},
		{"invalid profile", "invalid SSO account", func(map[string]any) {}},
		{"code expired", "SSO code rejected", func(map[string]any) {}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(token)
			var response map[string]any
			_ = json.Unmarshal(data, &response)
			tc.change(response)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.name == "code expired" {
					w.WriteHeader(400)
				} else if r.URL.Path == "/api/accounts/profile" {
					profile := snapshot.Profile
					if tc.name == "invalid profile" {
						profile.ID = ""
					}
					_ = json.NewEncoder(w).Encode(profile)
				} else {
					_ = json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()
			account, secret := email, password
			if tc.name == "wrong account" {
				account = "other@example.invalid"
			}
			if tc.name == "wrong password" {
				secret = "wrong password"
			}
			client, _ := NewClient(server.URL, account, testDevice)
			if session, err := client.LoginSSO(t.Context(), []byte(secret), testSSOCode()); err == nil || session != nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal("unexpected login result", err)
			}
		})
	}
}

func TestSSOCodeValidation(t *testing.T) {
	for _, change := range []func(*SSOCode){
		func(s *SSOCode) { s.Code = nil },
		func(s *SSOCode) { s.Code = []byte(strings.Repeat("c", 4097)) },
		func(s *SSOCode) { s.Verifier = []byte(strings.Repeat("v", 42)) },
		func(s *SSOCode) { s.Verifier = []byte(strings.Repeat("v", 129)) },
		func(s *SSOCode) { s.Verifier[0] = '\n' },
		func(s *SSOCode) { s.Verifier[0] = '%' },
		func(s *SSOCode) { s.RedirectURI = "https://attacker.invalid" },
		func(s *SSOCode) { s.RedirectURI = "http://localhost:8065/elsewhere" },
		func(s *SSOCode) { s.RedirectURI = "http://localhost:8071" },
	} {
		code := testSSOCode()
		change(&code)
		if code.Validate() == nil {
			t.Fatal("invalid SSO parameters accepted")
		}
	}
	client, _ := NewClient("https://example.invalid", "", testDevice)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.LoginSSO(ctx, []byte("password"), testSSOCode()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
