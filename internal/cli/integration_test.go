//go:build integration

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDaemonAgainstSnapshot(t *testing.T) {
	server := os.Getenv("BWB_INTEGRATION_URL")
	if server == "" {
		t.Fatal("run through dev/integration/run.py")
	}
	data, err := os.ReadFile("../../testdata/integration/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Email, Password string
		Items           map[string]string
		Organizations   struct {
			RestrictedEmail string `json:"restricted_email"`
			Entries         []struct{ Items []struct{ ID string } }
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "bwb")
	build := exec.Command("go", "build", "-o", binary, "../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	// Keep the Unix socket path within the platform limit.
	state, err := os.MkdirTemp("/tmp", "bwb-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)
	instance := uuid.NewString()
	target, err := url.Parse(server)
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	var requests atomic.Int64
	var offline atomic.Bool
	var pauseSync atomic.Bool
	var rejectRefresh atomic.Bool
	syncStarted := make(chan struct{}, 1)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if rejectRefresh.Load() && r.URL.Path == "/identity/connect/token" {
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
			form, _ := url.ParseQuery(string(body))
			if form.Get("grant_type") == "refresh_token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if pauseSync.Load() && r.URL.Path == "/api/sync" {
			syncStarted <- struct{}{}
			<-r.Context().Done()
			return
		}
		if offline.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	defer endpoint.Close()
	run := func(password string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, args...)
		command.Env = append(os.Environ(), "XDG_STATE_HOME="+state, "BWB_INSTANCE="+instance)
		command.Stdin = strings.NewReader(password)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if err != nil {
			return stderr.String(), err
		}
		if stderr.Len() != 0 {
			t.Fatalf("unexpected stderr: %s", stderr.String())
		}
		return stdout.String(), nil
	}
	must := func(password, want string, args ...string) {
		t.Helper()
		got, err := run(password, args...)
		if err != nil || got != want {
			t.Fatalf("%v: output %q, want %q, error %v", args, got, want, err)
		}
	}
	defer func() {
		for _, profile := range []string{"default", "other", "restricted"} {
			if output, err := run("", "--profile", profile, "logout"); err != nil {
				t.Errorf("credential cleanup: %v %s", err, output)
			}
		}
		if output, err := run("", "daemon", "stop"); err != nil {
			t.Errorf("daemon cleanup: %v %s", err, output)
		}
	}()
	must("", "stopped\n", "status")
	entries, err := os.ReadDir(state)
	if err != nil || len(entries) != 0 {
		t.Fatal("status created daemon state", err)
	}
	login := []string{"login", "--server", endpoint.URL, "--email", manifest.Email, "--password-stdin"}
	must(manifest.Password+"\n", "", login...)
	must("", "", "daemon", "update")
	must("", "unlocked\n", "status")
	must("", "logged_out\n", "--profile", "other", "status")
	list := func(kind string, flags ...string) []vaultapi.ListEntry {
		t.Helper()
		args := append([]string{kind, "list", "--format", "json"}, flags...)
		output, err := run("", args...)
		if err != nil {
			t.Fatalf("list %v: %v %s", args, err, output)
		}
		var result struct {
			Version int
			Data    []vaultapi.ListEntry
		}
		if err := json.Unmarshal([]byte(output), &result); err != nil || result.Version != 1 || result.Data == nil {
			t.Fatal("invalid list JSON", err, output)
		}
		for _, secret := range []string{manifest.Password, "DEMO-only-password-1!", "DEMO-org-hidden!", "DEMO-org-multi!", "DEMO-0001"} {
			if strings.Contains(output, secret) {
				t.Fatal("list leaked secret")
			}
		}
		return result.Data
	}
	before := requests.Load()
	offline.Store(true)
	all := list("item")
	expectedIDs := map[string]bool{}
	for _, id := range manifest.Items {
		expectedIDs[id] = true
	}
	for _, org := range manifest.Organizations.Entries {
		for _, item := range org.Items {
			expectedIDs[item.ID] = true
		}
	}
	if len(all) != len(expectedIDs) {
		t.Fatal("list missing items")
	}
	for _, entry := range all {
		if !expectedIDs[entry.ID] {
			t.Fatal("unexpected list item")
		}
		delete(expectedIDs, entry.ID)
	}
	if len(list("organization")) != 2 || len(list("collection")) != 4 || len(list("folder")) != 2 {
		t.Fatal("discovery metadata missing")
	}
	var shared vaultapi.ListEntry
	for _, entry := range all {
		if entry.ID == manifest.Organizations.Entries[0].Items[0].ID {
			shared = entry
		}
	}
	if shared.Organization == nil || shared.Folder == nil || len(shared.Collections) != 2 {
		t.Fatal("list lost folder/collection relationships")
	}
	selected := list("item", "--organization", shared.Organization.ID, "--collection", shared.Collections[0].ID, "--folder", shared.Folder.ID)
	if len(selected) != 1 || selected[0].ID != shared.ID {
		t.Fatal("combined filter mismatch")
	}
	if len(list("collection", "--organization", shared.Organization.ID)) != 2 {
		t.Fatal("collection organization filter mismatch")
	}
	if len(list("item", "--search", "DEMO shared LOGIN")) != 4 {
		t.Fatal("case-insensitive search failed")
	}
	if len(list("item", "--search", "no-such-item")) != 0 {
		t.Fatal("empty search failed")
	}
	descending := list("item", "--sort", "type,name", "--reverse")
	if len(descending) != len(all) {
		t.Fatal("sort lost items")
	}
	if output, err := run("", "item", "list", "--group-by", "collection"); err != nil || !strings.Contains(output, "collection: Shared [") || strings.Count(output, shared.ID) != 2 {
		t.Fatal("grouped table failed", err)
	}

	for _, tc := range []struct{ fixtureID, path, want string }{
		{"20000000-0000-4000-8000-000000000001", "login/password", "DEMO-only-password-1!"},
		{"20000000-0000-4000-8000-000000000005", "login/password", ""},
		{"20000000-0000-4000-8000-000000000006", "notes", "DEMO-0001\nDEMO-0002\n"},
	} {
		must("", tc.want, "read", manifest.Items[tc.fixtureID], tc.path)
	}
	for _, org := range manifest.Organizations.Entries {
		must("", "DEMO-org-multi!", "read", org.Items[0].ID, "login/password")
	}
	if requests.Load() != before {
		t.Fatal("cached reads made HTTP requests")
	}
	if _, err := run("", "sync"); err == nil {
		t.Fatal("offline sync succeeded")
	}
	id := manifest.Items["20000000-0000-4000-8000-000000000001"]
	must("", "DEMO-only-password-1!", "read", id, "login/password")
	offline.Store(false)
	must("", "", "sync")
	before = requests.Load()
	if _, err := run(manifest.Password+"\n", "login", "--server", endpoint.URL, "--email", "other@example.invalid", "--password-stdin"); err == nil {
		t.Fatal("profile rebound without logout")
	}
	if requests.Load() != before {
		t.Fatal("identity conflict reached network")
	}
	if _, err := run("wrong-password\n", login...); err == nil {
		t.Fatal("bad login succeeded")
	}
	must("", "DEMO-only-password-1!", "read", id, "login/password")
	must(manifest.Password+"\n", "", append([]string{"--profile", "other"}, login...)...)
	// Lock must cancel an in-flight replacement login and forbid late publication.
	pauseSync.Store(true)
	loginDone := make(chan error, 1)
	go func() { _, err := run(manifest.Password+"\n", login...); loginDone <- err }()
	select {
	case <-syncStarted:
	case <-time.After(10 * time.Second):
		t.Fatal("login did not reach sync")
	}
	must("", "", "lock")
	for _, org := range manifest.Organizations.Entries {
		if output, err := run("", "read", org.Items[0].ID, "login/password"); err == nil || strings.Contains(output, "DEMO-org-multi!") {
			t.Fatal("organization read survived lock", err)
		}
	}
	select {
	case err := <-loginDone:
		if err == nil {
			t.Fatal("login survived lock")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("lock failed to cancel login")
	}
	pauseSync.Store(false)
	must("", "locked\n", "status")
	if output, err := run("", "read", id, "login/password"); err == nil || !strings.Contains(output, "vault locked") {
		t.Fatal("read survived lock", err, output)
	}
	must("", "DEMO-only-password-1!", "--profile", "other", "read", id, "login/password")
	must("", "", "daemon", "stop")
	must("", "stopped\n", "status")
	// read starts a fresh daemon, which must have no previous profile keys.
	if output, err := run("", "--profile", "other", "read", id, "login/password"); err == nil || !strings.Contains(output, "vault locked") {
		t.Fatal("session survived restart", err, output)
	}
	offline.Store(true)
	before = requests.Load()
	if _, err := run("wrong-password\n", "unlock", "--password-stdin"); err == nil {
		t.Fatal("wrong password unlocked persisted cache")
	}
	must("", "locked\n", "status")
	must(manifest.Password+"\n", "", "unlock", "--password-stdin")
	must("", "DEMO-only-password-1!", "read", id, "login/password")
	for _, org := range manifest.Organizations.Entries {
		must("", "DEMO-org-multi!", "read", org.Items[0].ID, "login/password")
	}
	if len(list("item")) != len(all) {
		t.Fatal("offline restart list failed")
	}
	if requests.Load() != before {
		t.Fatal("offline unlock contacted server")
	}
	must("", "", "lock")
	must(manifest.Password+"\n", "", "unlock", "--password-stdin")
	offline.Store(false)
	before = requests.Load()
	must("", "", "sync")
	if requests.Load() < before+2 {
		t.Fatal("restart did not refresh using persisted token")
	}
	must(manifest.Password+"\n", "", "--profile", "restricted", "login", "--server", endpoint.URL, "--email", manifest.Organizations.RestrictedEmail, "--password-stdin")
	for _, org := range manifest.Organizations.Entries {
		must("", "DEMO-org-multi!", "--profile", "restricted", "read", org.Items[0].ID, "login/password")
		if output, err := run("", "--profile", "restricted", "read", org.Items[1].ID, "login/password"); err == nil || !strings.Contains(output, "restricted") || strings.Contains(output, "DEMO-org-hidden!") {
			t.Fatal("restricted secret exposed", err)
		}
	}
	restrictedItems := list("item", "--profile", "restricted")
	if len(restrictedItems) != 6 {
		t.Fatal("restricted profile list leaked personal items")
	}
	must("", "", "--profile", "restricted", "logout")
	// Exercise the actual controlling terminal; stdout remains a pipe.
	terminal := func(scenario, want string, args ...string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		script := exec.CommandContext(ctx, "python3", append([]string{"../../dev/integration/terminal.py", scenario, binary}, args...)...)
		script.Env = append(os.Environ(), "XDG_STATE_HOME="+state, "BWB_INSTANCE="+instance, "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_EXPECT="+want)
		if output, err := script.CombinedOutput(); err != nil {
			t.Fatalf("TTY %s: %v: %s", scenario, err, output)
		}
	}
	must("", "", "lock")
	terminal("no_prompt", "", "--non-interactive", "item", "list")
	terminal("no_prompt", "", "--non-interactive", "read", id, "login/password")
	for _, scenario := range []string{"wrong", "long", "cancel", "signal", "lock", "cancel_waiters"} {
		terminal(scenario, "", "read", id, "login/password")
		must("", "locked\n", "status")
	}
	offline.Store(true)
	before = requests.Load()
	terminal("concurrent", "DEMO-only-password-1!", "read", id, "login/password")
	listOutput, err := run("", "item", "list", "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	must("", "", "lock")
	terminal("success", listOutput, "item", "list", "--format", "json")

	if requests.Load() != before {
		t.Fatal("automatic unlock contacted server")
	}
	terminal("no_prompt", "", "sync")
	offline.Store(false)
	// Offline unlock has no access token, so sync must refresh, then recover a rejected refresh via login.
	must("", "", "lock")
	must(manifest.Password+"\n", "", "unlock", "--password-stdin")
	rejectRefresh.Store(true)
	if _, err := run("", "sync"); err == nil {
		t.Fatal("rejected refresh did not require login")
	}
	before = requests.Load()
	must("", "DEMO-only-password-1!", "read", id, "login/password")
	if requests.Load() != before {
		t.Fatal("cached read checked expired authentication")
	}
	terminal("success", "", "sync")
	rejectRefresh.Store(false)
	must("", "unlocked\n", "status")
	// A second restart proves the latest credential remains usable.
	must("", "", "daemon", "stop")
	must(manifest.Password+"\n", "", "unlock", "--password-stdin")
	must("", "", "sync")
	must("", "", "logout")
	must("", "logged_out\n", "status")
	if _, err := run(manifest.Password+"\n", "unlock", "--password-stdin"); err == nil {
		t.Fatal("logout preserved local unlock")
	}
	must(manifest.Password+"\n", "", "--profile", "other", "unlock", "--password-stdin")
	must("", "DEMO-only-password-1!", "--profile", "other", "read", id, "login/password")
	must("", "", "daemon", "stop")
	if _, err := run(manifest.Password+"\n", "unlock", "--password-stdin"); err == nil {
		t.Fatal("logout did not survive restart")
	}
}
