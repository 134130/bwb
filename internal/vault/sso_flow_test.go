package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
	"github.com/134130/go-daemonkit/secrets"
	"github.com/google/uuid"
)

type ssoFlowFixture struct {
	service                *Service
	secrets                *faultyCredentials
	plan                   LoginPlanRequest
	password               string
	apiKey                 bitwarden.APIKey
	noSSO                  atomic.Bool
	exchanges, discoveries atomic.Int32
}

func newSSOFlow(t *testing.T) *ssoFlowFixture {
	t.Helper()
	f := &ssoFlowFixture{}
	data, err := os.ReadFile("../../testdata/apikey.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Email, Password, ClientID, ClientSecret string
		Token                                   map[string]any
		Sync                                    json.RawMessage
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	fixture.Token["refresh_token"] = "flow-refresh"
	var snapshot struct{ Profile json.RawMessage }
	_ = json.Unmarshal(fixture.Sync, &snapshot)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/organizations/domain/sso/verified":
			f.discoveries.Add(1)
			if f.noSSO.Load() {
				_, _ = w.Write([]byte(`{"data":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"organizationIdentifier":"work"}]}`))
		case "/identity/connect/token":
			_ = r.ParseForm()
			if r.Form.Get("grant_type") == "client_credentials" {
				if r.Form.Get("client_id") != fixture.ClientID || r.Form.Get("client_secret") != fixture.ClientSecret {
					t.Error("wrong API key")
				}
				_ = json.NewEncoder(w).Encode(fixture.Token)
				return
			}
			f.exchanges.Add(1)
			if r.Form.Get("client_id") != "cli" || r.Form.Get("grant_type") != "authorization_code" {
				t.Error("wrong login flow")
			}
			_ = json.NewEncoder(w).Encode(fixture.Token)
		case "/api/accounts/profile":
			_, _ = w.Write(snapshot.Profile)
		case "/api/sync":
			_, _ = w.Write(fixture.Sync)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(server.Close)
	f.service, f.secrets = persistentService(t)
	if _, err := f.service.load(t.Context(), "default"); err != nil {
		t.Fatal(err)
	}
	_, _, done, err := f.service.begin(t.Context(), "default", true)
	if err != nil {
		t.Fatal(err)
	}
	done()
	t.Cleanup(func() { _ = f.service.Shutdown(context.Background()) })
	f.plan = LoginPlanRequest{Profile: "default", Server: server.URL, Email: fixture.Email}
	f.password = fixture.Password
	f.apiKey = bitwarden.APIKey{ClientID: fixture.ClientID, ClientSecret: []byte(fixture.ClientSecret)}
	return f
}

func flowCode() bitwarden.SSOCode {
	return bitwarden.SSOCode{Code: []byte("flow-code"), Verifier: []byte(strings.Repeat("v", 43)), RedirectURI: "http://localhost:8065", Identifier: "work"}
}

func flowAuthRequest(f *ssoFlowFixture, plan LoginPlanResponse) AuthRequest {
	return AuthRequest{Profile: f.plan.Profile, Action: "sso", Server: plan.Server, Email: plan.Email, Identifier: plan.Identifier, Epoch: plan.Epoch, Version: plan.Version}
}

func TestAPIKeyProfileCanLoginSSO(t *testing.T) {
	for _, mode := range []string{"discovered", "explicit", "one shot"} {
		t.Run(mode, func(t *testing.T) {
			f := newSSOFlow(t)
			s := f.service
			login := LoginRequest{Profile: "default", Server: f.plan.Server, Password: []byte(f.password)}
			if _, err := s.LoginAPIKey(t.Context(), LoginAPIKeyRequest{LoginRequest: login, APIKey: f.apiKey}); err != nil {
				t.Fatal(err)
			}
			old, err := s.load(t.Context(), "default")
			if err != nil {
				t.Fatal(err)
			}
			oldSession := s.profiles["default"].session
			unchanged := func() {
				t.Helper()
				saved, err := s.load(t.Context(), "default")
				if err != nil || saved.AuthMethod != "apikey" || saved.Credential != old.Credential || s.profiles["default"].session != oldSession {
					t.Fatal("unfinished SSO replaced API key login", err)
				}
				auth, err := s.loadCredential(t.Context(), saved)
				defer auth.clear()
				if err != nil || auth.ClientID != f.apiKey.ClientID || len(auth.ClientSecret) == 0 {
					t.Fatal("previous API key no longer usable", err)
				}
			}
			f.noSSO.Store(true)
			_, err = s.PlanLogin(t.Context(), LoginPlanRequest{Profile: "default"})
			var needed *AuthRequired
			if !errors.As(err, &needed) || needed.Method != "apikey" {
				t.Fatal("API key profile fell back to password", err)
			}
			if mode == "one shot" {
				login.Password = []byte(f.password)
				if _, err := s.LoginSSO(t.Context(), LoginSSORequest{LoginRequest: login, SSOCode: flowCode()}); err != nil {
					t.Fatal(err)
				}
			} else {
				f.noSSO.Store(mode == "explicit")
				for _, outcome := range []string{"cancel", "keychain failure", "success"} {
					plan, err := s.PlanLogin(t.Context(), LoginPlanRequest{Profile: "default", SSO: mode == "explicit"})
					if err != nil || plan.Method != "sso" || plan.Server != f.plan.Server || plan.Email != f.plan.Email {
						t.Fatal("saved API key blocked SSO discovery", plan, err)
					}
					reservation, err := s.BeginAuth(t.Context(), flowAuthRequest(f, plan))
					if err != nil {
						t.Fatal(err)
					}
					end := AuthEndRequest{Profile: "default", ID: reservation.ID}
					if _, err := s.ExchangeSSO(t.Context(), ExchangeSSORequest{AuthEndRequest: end, SSOCode: flowCode()}); err != nil {
						t.Fatal(err)
					}
					result, err := s.FinishSSO(t.Context(), UnlockRequest{Profile: "default", AuthID: end.ID, Password: []byte("wrong")})
					if err != nil || !result.RetryPassword {
						t.Fatal("wrong password did not allow retry", err)
					}
					unchanged()
					if outcome == "cancel" {
						if _, err := s.EndAuth(t.Context(), end); err != nil {
							t.Fatal(err)
						}
						unchanged()
						continue
					}
					if outcome == "keychain failure" {
						f.secrets.setErr = secrets.ErrAccess
					}
					_, err = s.FinishSSO(t.Context(), UnlockRequest{Profile: "default", AuthID: end.ID, Password: []byte(f.password)})
					f.secrets.setErr = nil
					if outcome == "keychain failure" {
						if !errors.Is(err, secrets.ErrAccess) {
							t.Fatal(err)
						}
						unchanged()
					} else if err != nil {
						t.Fatal(err)
					}
				}
			}
			saved, err := s.load(t.Context(), "default")
			if err != nil || saved.AuthMethod != "sso" || saved.Version != 3 || saved.Credential == old.Credential || oldSession.HasCredentials() {
				t.Fatal("SSO did not replace API key session", err)
			}
			auth, err := s.loadCredential(t.Context(), saved)
			defer auth.clear()
			if err != nil || string(auth.Token) != "flow-refresh" || auth.ClientID != "" || len(auth.ClientSecret) != 0 {
				t.Fatal("SSO credential retained API key", err)
			}
			if _, err := s.Logout(t.Context(), ProfileRequest{Profile: "default"}); err != nil || len(f.secrets.values) != 0 {
				t.Fatal("logout retained credentials", err)
			}
		})
	}
}

func TestSSOFlowRetryAndRememberIdentity(t *testing.T) {
	f := newSSOFlow(t)
	s := f.service
	plan, err := s.PlanLogin(t.Context(), f.plan)
	if err != nil || plan.Method != "sso" || plan.Identifier != "work" || plan.Discovery != "matched" {
		t.Fatal(plan, err)
	}
	reservation, err := s.BeginAuth(t.Context(), flowAuthRequest(f, plan))
	if err != nil {
		t.Fatal(err)
	}
	req := ExchangeSSORequest{AuthEndRequest: AuthEndRequest{Profile: "default", ID: reservation.ID}, SSOCode: flowCode()}
	account, err := s.ExchangeSSO(t.Context(), req)
	if err != nil || account.Email != f.plan.Email {
		t.Fatal(account, err)
	}
	if !bytes.Equal(req.Code, make([]byte, len(req.Code))) || !bytes.Equal(req.Verifier, make([]byte, len(req.Verifier))) {
		t.Fatal("exchange retained RPC secrets")
	}
	saved, err := s.load(t.Context(), "default")
	if err != nil || saved.Cache != nil || len(f.secrets.values) != 0 {
		t.Fatal("exchange persisted an unfinished login", err)
	}
	pending := s.profiles["default"].auth.pending
	password := []byte("wrong")
	result, err := s.FinishSSO(t.Context(), UnlockRequest{Profile: "default", AuthID: reservation.ID, Password: password})
	if err != nil || !result.RetryPassword || !pending.HasCredentials() || !bytes.Equal(password, make([]byte, len(password))) {
		t.Fatal("wrong password discarded authentication", err)
	}
	result, err = s.FinishSSO(t.Context(), UnlockRequest{Profile: "default", AuthID: reservation.ID, Password: []byte(f.password)})
	if err != nil || result.RetryPassword || f.exchanges.Load() != 1 || pending.HasCredentials() {
		t.Fatal("retry did not reuse and dispose of pending authentication", err)
	}
	remembered, err := s.PlanLogin(t.Context(), LoginPlanRequest{Profile: "default"})
	if err != nil || remembered.Method != "sso" || remembered.Server != f.plan.Server || remembered.Email != f.plan.Email || remembered.Identifier != "work" || f.discoveries.Load() != 1 {
		t.Fatal("saved login was not reused", remembered, err)
	}
	if _, err := s.Lock(t.Context(), ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Unlock(t.Context(), UnlockRequest{Profile: "default", Password: []byte(f.password)}); err != nil {
		t.Fatal(err)
	}
	saved, err = s.load(t.Context(), "default")
	if err != nil || saved.Cache.SSOIdentifier != "work" {
		t.Fatal("offline unlock lost identifier", err)
	}
}

func TestSSOPlanDistinguishesUnavailableDiscovery(t *testing.T) {
	s, _ := persistentService(t)
	for _, tc := range []struct {
		name, body, discovery string
		status                int
		fail                  bool
	}{
		{"none", `{"data":[]}`, "not_found", 200, false},
		{"unsupported", ``, "unsupported", 404, false},
		{"outage", ``, "", 503, true},
		{"malformed", `{}`, "", 200, true},
		{"multiple", `{"data":[{"organizationIdentifier":"a"},{"organizationIdentifier":"b"}]}`, "", 200, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			plan, err := s.PlanLogin(t.Context(), LoginPlanRequest{Profile: "default", Server: server.URL, Email: "user@example.invalid"})
			if (err != nil) != tc.fail || !tc.fail && (plan.Method != "password" || plan.Discovery != tc.discovery) {
				t.Fatal(plan, err)
			}
			forced, err := s.PlanLogin(t.Context(), LoginPlanRequest{Profile: "default", Server: server.URL, SSO: true})
			if err != nil || forced.Method != "sso" {
				t.Fatal("explicit SSO depended on domain discovery", err)
			}
		})
	}
}

func TestSSORejectsConfigIdentityChangeDuringBrowserLogin(t *testing.T) {
	f := newSSOFlow(t)
	s := f.service
	plan, err := s.PlanLogin(t.Context(), f.plan)
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := s.BeginAuth(t.Context(), flowAuthRequest(f, plan))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ExchangeSSO(t.Context(), ExchangeSSORequest{AuthEndRequest: AuthEndRequest{Profile: "default", ID: reservation.ID}, SSOCode: flowCode()}); err != nil {
		t.Fatal(err)
	}
	pending := s.profiles["default"].auth.pending
	if err := s.ApplyConfig(t.Context(), &config.File{Profiles: map[string]config.Profile{"default": {Email: "different@example.invalid"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FinishSSO(t.Context(), UnlockRequest{Profile: "default", AuthID: reservation.ID, Password: []byte(f.password)}); err == nil {
		t.Fatal("login ignored changed configured identity")
	}
	saved, err := s.load(t.Context(), "default")
	if err != nil || saved.Cache != nil || len(f.secrets.values) != 0 || pending.HasCredentials() {
		t.Fatal("failed login retained credentials", err)
	}
}

func TestSSOReservationCancellationAndExpiration(t *testing.T) {
	for _, reason := range []string{"lock", "logout", "owner disconnect", "timeout", "shutdown", "claimed timeout"} {
		t.Run(reason, func(t *testing.T) {
			f := newSSOFlow(t)
			s := f.service
			client, _ := bitwarden.NewClient(f.plan.Server, f.plan.Email, uuid.NewString())
			pending, err := client.ExchangeSSO(t.Context(), flowCode())
			if err != nil {
				t.Fatal(err)
			}
			defer pending.Close()
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				req := AuthRequest{Profile: "default", Action: "sso", Server: f.plan.Server, Email: f.plan.Email, Identifier: "work"}
				first, err := s.BeginAuth(ctx, req)
				if err != nil {
					t.Fatal(err)
				}
				s.mu.Lock()
				s.profiles["default"].auth.pending = pending
				s.mu.Unlock()
				waiting := make(chan error, 1)
				go func() { _, err := s.BeginAuth(ctx, req); waiting <- err }()
				synctest.Wait()
				var operation context.Context
				if reason == "claimed timeout" {
					var done func()
					var p *profile
					operation, p, done, err = s.begin(ctx, "default", true)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { done() }()
					if err := s.claimAuth(operation, p, first.ID, "sso"); err != nil {
						t.Fatal(err)
					}
					// FinishSSO owns a detached token while work is in flight.
					s.mu.Lock()
					p.auth.pending = nil
					s.mu.Unlock()
					time.Sleep(SSOTimeout)
					<-operation.Done()
					if operation.Err() == nil {
						t.Fatal("claimed operation outlived reservation")
					}
					pending.Close()
					done()
					done = func() {}
				} else {
					switch reason {
					case "lock":
						_, err = s.Lock(ctx, ProfileRequest{Profile: "default"})
					case "logout":
						_, err = s.Logout(ctx, ProfileRequest{Profile: "default"})
					case "shutdown":
						err = s.Shutdown(ctx)
					case "timeout":
						time.Sleep(SSOTimeout)
					case "owner disconnect":
						watchCtx, cancel := context.WithCancel(ctx)
						watched := make(chan struct{})
						go func() {
							defer close(watched)
							_, _ = s.WatchAuth(watchCtx, AuthEndRequest{Profile: "default", ID: first.ID})
						}()
						synctest.Wait()
						cancel()
						<-watched
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				synctest.Wait()
				if err := <-waiting; !errors.Is(err, ErrStateChanged) {
					t.Fatal("canceled waiter opened another browser", err)
				}
				if pending.HasCredentials() {
					t.Fatal("pending token survived cancellation")
				}
				if _, err := s.ExchangeSSO(ctx, ExchangeSSORequest{AuthEndRequest: AuthEndRequest{Profile: "default", ID: first.ID}, SSOCode: flowCode()}); !errors.Is(err, ErrStateChanged) {
					t.Fatal("late callback accepted", err)
				}
				if _, err := s.FinishSSO(ctx, UnlockRequest{Profile: "default", AuthID: first.ID, Password: []byte(f.password)}); !errors.Is(err, ErrStateChanged) {
					t.Fatal("late unlock accepted", err)
				}
			})
		})
	}
}

func TestSSOConcurrentWaiterReusesSuccessfulLogin(t *testing.T) {
	f := newSSOFlow(t)
	s := f.service
	synctest.Test(t, func(t *testing.T) {
		req := AuthRequest{Profile: "default", Action: "sso", Server: f.plan.Server, Email: f.plan.Email, Identifier: "work"}
		first, err := s.BeginAuth(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		waiting := make(chan AuthResponse, 1)
		go func() {
			r, err := s.BeginAuth(context.Background(), req)
			if err != nil {
				t.Error(err)
			}
			waiting <- r
		}()
		synctest.Wait()
		select {
		case <-waiting:
			t.Fatal("second browser owner elected")
		default:
		}
		s.mu.Lock()
		p := s.profiles["default"]
		p.authVersion++
		p.auth.succeeded = true
		s.finishAuth(p)
		s.mu.Unlock()
		synctest.Wait()
		if r := <-waiting; !r.Ready || r.ID != "" {
			t.Fatal("waiter did not reuse successful authentication")
		}
		watch, err := s.WatchAuth(context.Background(), AuthEndRequest{Profile: "default", ID: first.ID})
		if err != nil || !watch.Succeeded {
			t.Fatal(watch, err)
		}
	})
}
