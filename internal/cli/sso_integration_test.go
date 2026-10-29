//go:build integration

package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBrowserSSOThroughDaemon(t *testing.T) {
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
	fixture.Token["refresh_token"] = strings.Repeat("sso-refresh.", 512)
	rotated := strings.Repeat("sso-rotated.", 1024)
	var snapshot struct{ Profile json.RawMessage }
	if err := json.Unmarshal(fixture.Sync, &snapshot); err != nil {
		t.Fatal(err)
	}
	var exchanges, discoveries, refreshes atomic.Int32
	var failureStage atomic.Int32
	var challenge atomic.Value
	var expectedRefresh atomic.Value
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/organizations/domain/sso/verified":
			discoveries.Add(1)
			_, _ = w.Write([]byte(`{"data":[{"organizationIdentifier":"work"}]}`))
		case "/test/challenge":
			challenge.Store(r.URL.Query().Get("value"))
		case "/test/exchanges":
			_, _ = fmt.Fprint(w, exchanges.Load())
		case "/identity/connect/token":
			_ = r.ParseForm()
			if failureStage.Load() == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			if r.Form.Get("grant_type") != "client_credentials" && r.Form.Get("client_id") != "cli" || r.Form.Has("password") {
				t.Error("wrong SSO token request")
			}
			switch r.Form.Get("grant_type") {
			case "client_credentials":
				if r.Form.Get("client_id") != fixture.ClientID || r.Form.Get("client_secret") != fixture.ClientSecret {
					t.Error("wrong API key")
				}
			case "authorization_code":
				exchanges.Add(1)
				expectedRefresh.Store(fixture.Token["refresh_token"])
				digest := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				if base64.RawURLEncoding.EncodeToString(digest[:]) != challenge.Load() {
					t.Error("PKCE mismatch")
				}
			case "refresh_token":
				refreshes.Add(1)
				if r.Form.Get("refresh_token") != expectedRefresh.Load() {
					t.Error("stored refresh token differs from last issued token")
				}
				expectedRefresh.Store(rotated)
				response := maps.Clone(fixture.Token)
				response["refresh_token"] = rotated
				_ = json.NewEncoder(w).Encode(response)
				return
			default:
				t.Error("unexpected token grant")
			}
			_ = json.NewEncoder(w).Encode(fixture.Token)
		case "/api/accounts/profile":
			_, _ = w.Write(snapshot.Profile)
		case "/api/sync":
			switch failureStage.Load() {
			case 2:
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			case 3:
				_, _ = w.Write([]byte(`{}`))
				return
			}
			_, _ = w.Write(fixture.Sync)
		default:
			t.Error("unexpected endpoint", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer endpoint.Close()
	binary := filepath.Join(t.TempDir(), "bwb")
	if output, err := exec.Command("go", "build", "-o", binary, "../..").CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	state, err := os.MkdirTemp("/tmp", "bwb-sso-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(state, "config"))
	t.Setenv("BWB_CONFIG", "")
	t.Setenv("BWB_PROFILE", "")
	t.Setenv("BWB_INSTANCE", uuid.NewString())
	t.Setenv("BWB_TEST_PASSWORD", fixture.Password)
	t.Setenv("BWB_TEST_EXPECT", "")
	t.Setenv("BWB_TEST_SERVER", endpoint.URL)
	t.Setenv("BWB_TEST_EMAIL", fixture.Email)
	t.Setenv("BWB_TEST_SSO_CHECK", endpoint.URL+"/test/exchanges")
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal(err)
	}
	opener := t.TempDir()
	script := "#!" + python + "\n" + `import os, sys, urllib.parse, urllib.request
uri = urllib.parse.urlsplit(sys.argv[1])
query = urllib.parse.parse_qs(uri.fragment.split('?', 1)[1])
assert query['clientId'] == ['cli'] and query['identifier'] == ['work'] and query['email']
with urllib.request.urlopen(urllib.parse.urlunsplit((uri.scheme, uri.netloc, '/test/challenge', urllib.parse.urlencode({'value': query['codeChallenge'][0]}), ''))):
    pass
if os.environ.get('BWB_TEST_SSO_HOLD') == '1':
    sys.exit(0)
callback = query['redirectUri'][0] + '?' + urllib.parse.urlencode({'code': 'browser-code', 'state': query['state'][0] + '_identifier=work'})
with urllib.request.urlopen(callback):
    pass
`
	for _, name := range []string{"open", "xdg-open"} {
		if err := os.WriteFile(filepath.Join(opener, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", opener+string(os.PathListSeparator)+os.Getenv("PATH"))
	run := func(password string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, append([]string{"--profile", "work"}, args...)...)
		if slices.Contains(args, "--apikey") {
			command.Env = append(os.Environ(), "BWB_CLIENTID="+fixture.ClientID, "BWB_CLIENTSECRET="+fixture.ClientSecret)
		}
		command.Stdin = strings.NewReader(password)
		output, err := command.CombinedOutput()
		return string(output), err
	}
	must := func(password string, args ...string) {
		t.Helper()
		if output, err := run(password, args...); err != nil {
			t.Fatalf("%v: %v: %s", args, err, output)
		}
	}
	defer func() { must("", "logout"); must("", "daemon", "stop") }()
	login := func(scenario, count string, args ...string) {
		t.Helper()
		t.Setenv("BWB_TEST_SSO_EXCHANGES", count)
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, python, append([]string{"../../dev/integration/terminal.py", scenario, binary, "--profile", "work", "login"}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("SSO terminal: %v: %s", err, output)
		}
	}
	for _, args := range [][]string{
		{"login", "--non-interactive", "--password-stdin"},
		{"login", "--server", endpoint.URL, "--non-interactive", "--password-stdin"},
	} {
		if output, err := run(fixture.Password+"\n", args...); err == nil || !strings.Contains(output, "required for first login") {
			t.Fatal("noninteractive login did not explain missing identity", output, err)
		}
	}
	login("setup_cancel", "0")
	login("setup_signal", "0")
	login("sso_setup", "1")
	login("cancel", "1")
	login("signal", "1")
	if output, err := run("\n", "login", "--password-stdin"); err == nil || !strings.Contains(output, "invalid password length") || strings.Contains(output, "Complete SSO login") {
		t.Fatal("empty password input reached browser SSO", err, output)
	}
	t.Setenv("BWB_TEST_SSO_HOLD", "1")
	login("sso_browser_cancel", "1")
	t.Setenv("BWB_TEST_SSO_HOLD", "")
	if exchanges.Load() != 1 {
		t.Fatal("canceled password or browser prompt exchanged an SSO code")
	}
	must("", "daemon", "stop")
	must(fixture.Password+"\n", "unlock", "--password-stdin")
	must("", "sync")
	if refreshes.Load() != 1 {
		t.Fatal("restart lost SSO refresh client")
	}
	must("", "daemon", "stop")
	must(fixture.Password+"\n", "unlock", "--password-stdin")
	must("", "sync")
	if refreshes.Load() != 2 {
		t.Fatal("restart lost rotated long SSO token")
	}
	login("sso", "2")
	if discoveries.Load() != 1 || exchanges.Load() != 2 {
		t.Fatal("saved identity or single exchange lost")
	}
	output, err := run(fixture.Password+"\n", "login", "--non-interactive", "--password-stdin")
	if err == nil || !strings.Contains(output, "requires browser interaction") || exchanges.Load() != 2 {
		t.Fatal("noninteractive SSO was attempted", err)
	}
	if bytes.Contains([]byte(output), []byte(fixture.Password)) {
		t.Fatal("password leaked")
	}
	// Existing API key profiles must also reach discovery and browser SSO through plain login.
	must("", "logout")
	must(fixture.Password+"\n", "login", "--apikey", "--server", endpoint.URL, "--password-stdin")
	login("sso", "3")
	if discoveries.Load() != 2 || exchanges.Load() != 3 {
		t.Fatal("API key profile blocked SSO discovery")
	}
	must("", "daemon", "stop")
	must(fixture.Password+"\n", "unlock", "--password-stdin")
	must("", "sync")
	if refreshes.Load() != 3 {
		t.Fatal("API key profile did not retain replacement SSO credential")
	}
	for _, tc := range []struct {
		stage int32
		want  string
	}{
		{1, "vault API returned HTTP 503"},
		{2, "vault API returned HTTP 503"},
		{3, "vault account does not match session"},
	} {
		failureStage.Store(tc.stage)
		for range 3 {
			output, err := run(fixture.Password+"\n", "login", "--password-stdin")
			if err == nil || !strings.Contains(output, tc.want) {
				t.Fatalf("SSO failure at stage %d lost its cause: %v: %s", tc.stage, err, output)
			}
			must("", "status")
		}
	}
}
