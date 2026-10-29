//go:build integration

package cli

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
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
	info, err := buildinfo.ReadFile(binary)
	if err != nil || !strings.HasPrefix(info.Main.Version, "v") {
		t.Fatal("build has no VCS version", err)
	}
	version := info.Main.Version
	// Keep the Unix socket path within the platform limit.
	state, err := os.MkdirTemp("/tmp", "bwb-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(state, "config"))
	t.Setenv("BWB_CONFIG", "")
	t.Setenv("BWB_PROFILE", "")
	// Replace OS clipboard commands before starting any CLI/TUI child.
	clipboardDir := t.TempDir()
	clipboardFile := filepath.Join(clipboardDir, "value")
	for _, name := range []string{"pbcopy", "xclip", "wl-copy", "wl-paste"} {
		if err := os.WriteFile(filepath.Join(clipboardDir, name), []byte("#!/bin/sh\nif [ -e \"$BWB_CLIPBOARD_FILE.fail\" ]; then exit 1; fi\ncat > \"$BWB_CLIPBOARD_FILE\"\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", clipboardDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BWB_CLIPBOARD_FILE", clipboardFile)

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
	var rejectAPIKey, expireAPIKey atomic.Bool
	var apiKeyRequests atomic.Int64
	syncStarted := make(chan struct{}, 1)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/identity/connect/token" {
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
			form, _ := url.ParseQuery(string(body))
			if form.Get("grant_type") == "client_credentials" {
				apiKeyRequests.Add(1)
				if form.Has("password") || form.Has("username") || form.Has("refresh_token") {
					t.Error("API key login sent password credentials")
				}
				if rejectAPIKey.Load() {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			if rejectRefresh.Load() && form.Get("grant_type") == "refresh_token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if r.URL.Path == "/api/sync" && expireAPIKey.CompareAndSwap(true, false) {
			w.WriteHeader(http.StatusUnauthorized)
			return
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
		if slices.Contains(args, "--apikey") {
			command.Env = append(command.Env, "BWB_CLIENTID="+os.Getenv("BWB_INTEGRATION_CLIENTID"), "BWB_CLIENTSECRET="+os.Getenv("BWB_INTEGRATION_CLIENTSECRET"))
		}
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
		for _, profile := range []string{"default", "other", "restricted", "timed", "steady", "api", "fresh"} {
			if output, err := run("", "--profile", profile, "logout"); err != nil {
				t.Errorf("credential cleanup: %v %s", err, output)
			}
		}
		if output, err := run("", "daemon", "stop"); err != nil {
			t.Errorf("daemon cleanup: %v %s", err, output)
		}
	}()
	must("", "stopped\n", "status")
	must("", "bwb version "+version+"\n", "--version")
	must("", "bwb version "+version+"\n", "--config", "missing.toml", "version")
	must("", "stopped\n", "daemon", "status")
	must("", "stopped\n", "status", "--verbose")
	entries, err := os.ReadDir(state)
	if err != nil || len(entries) != 0 {
		t.Fatal("status/version created daemon state", err)
	}
	login := []string{"login", "--server", endpoint.URL, "--email", manifest.Email, "--password-stdin"}
	must(manifest.Password+"\n", "", login...)
	must("", "", "daemon", "update")
	must("", "unlocked\n", "status")
	otpID := manifest.Items["20000000-0000-4000-8000-000000000009"]
	if code, err := run("", "--non-interactive", "totp", otpID); err != nil || !regexp.MustCompile("^[0-9]{6}$").MatchString(code) {
		t.Fatal("one-shot TOTP", err)
	}

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
		{"20000000-0000-4000-8000-000000000001", "password", "DEMO-only-password-1!"},
		{"20000000-0000-4000-8000-000000000005", "password", ""},
		{"20000000-0000-4000-8000-000000000006", "notes", "DEMO-0001\nDEMO-0002\n"},
	} {
		must("", tc.want, "item", "read", manifest.Items[tc.fixtureID], "--field", tc.path)
	}
	for _, org := range manifest.Organizations.Entries {
		must("", "DEMO-org-multi!", "item", "read", org.Items[0].ID, "--field", "password")
	}
	// Details discover field keys without returning hidden values; explicit selection returns raw data.
	if output, err := run("", "item", "read", "GitHub (demo)"); err != nil || !strings.Contains(output, "developer@example.invalid") || !strings.Contains(output, "--field") || strings.Contains(output, "DEMO-only-password-1!") {
		t.Fatal("item details or masking failed", err)
	}
	must("", "DEMO-only-password-1!", "item", "read", "github (DEMO)", "--field", "password")
	if output, err := run("", "item", "read", "GitHub (demo)", "--format", "json"); err != nil || strings.Contains(output, "DEMO-only-password-1!") || !strings.Contains(output, `"state":"hidden"`) {
		t.Fatal("JSON details revealed a hidden value", err)
	}
	if output, err := run("", "item", "read", "GitHub (demo)", "--reveal", "--format", "json"); err != nil || !strings.Contains(output, "DEMO-only-password-1!") {
		t.Fatal("explicit reveal failed", err)
	}
	if output, err := run("", "item", "read", "Demo Shared Login"); err == nil || !strings.Contains(output, "use an ID") {
		t.Fatal("duplicate name was silently selected", err)
	}
	if output, err := run("", "item", "read", "GitHub (demo)", "--field", "login/password"); err == nil || !strings.Contains(output, "available fields") {
		t.Fatal("legacy field path accepted or no discovery guidance", err)
	}
	if _, err := run("", "read", "unused", "login/password"); err == nil {
		t.Fatal("removed root read command still exists")
	}
	code, err := run("", "item", "totp", "OTP (public test seed)")
	if err != nil || len(code) != 6 || strings.Trim(code, "0123456789") != "" {
		t.Fatal("TOTP command did not return a code", err)
	}
	must("", "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "item", "read", "OTP (public test seed)", "--field", "totp-secret")
	if requests.Load() != before {
		t.Fatal("cached reads made HTTP requests")
	}
	if _, err := run("", "sync"); err == nil {
		t.Fatal("offline sync succeeded")
	}
	id := manifest.Items["20000000-0000-4000-8000-000000000001"]
	copyCommand := exec.Command(binary, "item", "read", id, "--field", "password", "--copy")
	copyCommand.Env = append(os.Environ(), "XDG_STATE_HOME="+state, "BWB_INSTANCE="+instance)
	var copyOut, copyErr bytes.Buffer
	copyCommand.Stdout, copyCommand.Stderr = &copyOut, &copyErr
	if err := copyCommand.Run(); err != nil || copyOut.Len() != 0 || copyErr.String() != "Copied\n" {
		t.Fatal("CLI copy output", err)
	}
	copied, err := os.ReadFile(clipboardFile)
	if err != nil || string(copied) != "DEMO-only-password-1!" {
		t.Fatal("CLI copy mismatch", err)
	}
	if err := os.WriteFile(clipboardFile+".fail", nil, 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := run("", "item", "read", id, "--field", "password", "--copy"); err == nil || !strings.Contains(output, "clipboard unavailable") || strings.Contains(output, "DEMO-only") {
		t.Fatal("copy failure handling", err)
	}
	if err := os.Remove(clipboardFile + ".fail"); err != nil {
		t.Fatal(err)
	}

	must("", "DEMO-only-password-1!", "item", "read", id, "--field", "password")
	offline.Store(false)
	if _, err := run("", "item", "list", "--search", "GitHub", "--sync", "--format", "json"); err != nil {
		t.Fatal("search with sync failed", err)
	}
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
	must("", "DEMO-only-password-1!", "item", "read", id, "--field", "password")
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
		if output, err := run("", "item", "read", org.Items[0].ID, "--field", "password"); err == nil || strings.Contains(output, "DEMO-org-multi!") {
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
	if output, err := run("", "item", "read", id, "--field", "password"); err == nil || !strings.Contains(output, "vault locked") {
		t.Fatal("read survived lock", err, output)
	}
	must("", "DEMO-only-password-1!", "--profile", "other", "item", "read", id, "--field", "password")
	must("", "", "daemon", "stop")
	must("", "stopped\n", "status")
	// read starts a fresh daemon, which must have no previous profile keys.
	if output, err := run("", "--profile", "other", "item", "read", id, "--field", "password"); err == nil || !strings.Contains(output, "vault locked") {
		t.Fatal("session survived restart", err, output)
	}
	offline.Store(true)
	before = requests.Load()
	if _, err := run("wrong-password\n", "unlock", "--password-stdin"); err == nil {
		t.Fatal("wrong password unlocked persisted cache")
	}
	must("", "locked\n", "status")
	must(manifest.Password+"\n", "", "unlock", "--password-stdin")
	must("", "DEMO-only-password-1!", "item", "read", id, "--field", "password")
	for _, org := range manifest.Organizations.Entries {
		must("", "DEMO-org-multi!", "item", "read", org.Items[0].ID, "--field", "password")
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
		must("", "DEMO-org-multi!", "--profile", "restricted", "item", "read", org.Items[0].ID, "--field", "password")
		if output, err := run("", "--profile", "restricted", "item", "read", org.Items[1].ID, "--field", "password"); err == nil || !strings.Contains(output, "restricted") || strings.Contains(output, "DEMO-org-hidden!") {
			t.Fatal("restricted secret exposed", err)
		}
		if output, err := run("", "--profile", "restricted", "item", "read", org.Items[1].ID, "--reveal", "--format", "json"); err != nil || !strings.Contains(output, `"restriction":`) || !strings.Contains(output, `"fields":[]`) || strings.Contains(output, "DEMO-org-hidden!") {
			t.Fatal("reveal bypassed organization restriction", err)
		}
		if _, err := run("", "--profile", "restricted", "item", "totp", org.Items[1].ID); err == nil {
			t.Fatal("TOTP bypassed organization restriction")
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
		script.Env = append(os.Environ(), "XDG_STATE_HOME="+state, "BWB_INSTANCE="+instance, "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_EXPECT="+want, "BWB_TEST_SERVER="+endpoint.URL, "BWB_TEST_EMAIL="+manifest.Email)
		if output, err := script.CombinedOutput(); err != nil {
			t.Fatalf("TTY %s: %v: %s", scenario, err, output)
		}
	}
	tuiTerminal := func(scenario string, args ...string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if len(args) == 0 {
			args = []string{"tui"}
		}
		script := exec.CommandContext(ctx, "../../.dev/venv/bin/python", append([]string{"../../dev/integration/tui.py", binary, scenario}, args...)...)
		script.Env = append(os.Environ(), "XDG_STATE_HOME="+state, "BWB_INSTANCE="+instance, "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_SERVER="+endpoint.URL)
		if output, err := script.CombinedOutput(); err != nil {
			t.Fatalf("TUI %s: %v: %s", scenario, err, output)
		}
	}
	terminal("setup", "", "--profile", "fresh", "login")
	must("", "", "lock")
	tuiTerminal("cancel")
	tuiTerminal("explorer")
	must("", "", "lock")
	tuiTerminal("reprompt")
	protectedID := manifest.Items["20000000-0000-4000-8000-000000000014"]
	// A grant is explicit and item-scoped; raw RPCs cannot skip the daemon policy.
	t.Setenv("XDG_STATE_HOME", state)
	rpcClient, err := kit.Connect(t.Context(), kit.Options{AppID: "bwb", InstanceID: instance, Version: version})
	if err != nil {
		t.Fatal(err)
	}
	grant, err := rpcClient.Invoke(t.Context(), vaultapi.Reprompt, vaultapi.RepromptRequest{Profile: "default", Item: protectedID, Password: []byte(manifest.Password)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = rpcClient.Invoke(t.Context(), vaultapi.Item, vaultapi.ItemRequest{Profile: "default", Item: manifest.Items["20000000-0000-4000-8000-000000000015"], Field: "password", RepromptToken: grant.Token}); err == nil {
		t.Fatal("reprompt crossed item")
	}
	if code, err := rpcClient.Invoke(t.Context(), vaultapi.TOTP, vaultapi.TOTPRequest{Profile: "default", ID: protectedID, RepromptToken: grant.Token}); err != nil || code.Code == "" {
		t.Fatal("protected TOTP", err)
	}
	if _, err = rpcClient.Invoke(t.Context(), vaultapi.Sync, vaultapi.ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if _, err = rpcClient.Invoke(t.Context(), vaultapi.Item, vaultapi.ItemRequest{Profile: "default", Item: protectedID, Field: "password", RepromptToken: grant.Token}); err == nil {
		t.Fatal("sync retained grant")
	}
	rpcClient.Close()
	terminal("success", "DEMO-reprompt-password!", "item", "read", protectedID, "--field", "password")
	terminal("wrong", "", "item", "read", protectedID, "--field", "password")
	terminal("cancel", "", "item", "read", protectedID, "--field", "password")
	terminal("no_prompt", "", "--non-interactive", "item", "read", protectedID, "--field", "password")
	must("", "", "lock")
	tuiTerminal("otp", "totp", otpID)
	must("", "", "lock")
	terminal("no_prompt", "", "--non-interactive", "item", "list")
	terminal("no_prompt", "", "--non-interactive", "item", "read", id, "--field", "password")
	for _, scenario := range []string{"wrong", "long", "cancel", "signal", "lock", "cancel_waiters"} {
		terminal(scenario, "", "item", "read", id, "--field", "password")
		must("", "locked\n", "status")
	}
	offline.Store(true)
	before = requests.Load()
	terminal("concurrent", "DEMO-only-password-1!", "item", "read", id, "--field", "password")
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
	must("", "DEMO-only-password-1!", "item", "read", id, "--field", "password")
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
	must("", "DEMO-only-password-1!", "--profile", "other", "item", "read", id, "--field", "password")
	must("", "", "daemon", "stop")
	if _, err := run(manifest.Password+"\n", "unlock", "--password-stdin"); err == nil {
		t.Fatal("logout did not survive restart")
	}
	// Personal API keys must survive restart without relying on refresh tokens or launcher environment.
	apiLogin := []string{"--profile", "api", "login", "--server", endpoint.URL, "--apikey", "--password-stdin"}
	must(manifest.Password+"\n", "", apiLogin...)
	must("", "DEMO-only-password-1!", "--profile", "api", "item", "read", id, "--field", "password")
	apiBefore := apiKeyRequests.Load()
	expireAPIKey.Store(true)
	must("", "", "--profile", "api", "sync")
	if apiKeyRequests.Load() != apiBefore+1 {
		t.Fatal("API key 401 did not renew exactly once")
	}
	must("", "", "daemon", "stop")
	offline.Store(true)
	before = requests.Load()
	must(manifest.Password+"\n", "", "--profile", "api", "unlock", "--password-stdin")
	must("", "DEMO-only-password-1!", "--profile", "api", "item", "read", id, "--field", "password")
	if requests.Load() != before {
		t.Fatal("API key offline unlock contacted server")
	}
	offline.Store(false)
	apiBefore = apiKeyRequests.Load()
	must("", "", "--profile", "api", "sync")
	if apiKeyRequests.Load() != apiBefore+1 {
		t.Fatal("API key restart did not use keychain")
	}
	if _, err := run("wrong\n", apiLogin...); err == nil {
		t.Fatal("API login accepted wrong master password")
	}
	must("", "DEMO-only-password-1!", "--profile", "api", "item", "read", id, "--field", "password")
	must("", "", "--profile", "api", "lock")
	terminal("success", "", "--profile", "api", "unlock")
	rejectAPIKey.Store(true)
	if output, err := run("", "--profile", "api", "sync"); err == nil || !strings.Contains(output, "login --apikey") {
		t.Fatal("API key revocation not reported", err)
	}
	terminal("no_prompt", "", "--profile", "api", "sync")
	must("", "DEMO-only-password-1!", "--profile", "api", "item", "read", id, "--field", "password")
	rejectAPIKey.Store(false)
	terminal("apikey", "", "--profile", "api", "login", "--apikey", "--server", endpoint.URL)
	must("", "", "--profile", "api", "logout")
	must("", "logged_out\n", "--profile", "api", "status")
	if _, err := run(manifest.Password+"\n", "--profile", "api", "unlock", "--password-stdin"); err == nil {
		t.Fatal("API key logout retained unlock")
	}
	terminal("apikey_setup", "", "--profile", "api", "login", "--apikey")
	must("", "", "--profile", "api", "logout")

	// Daemon loads config at startup; edits take effect only through reload.
	must("", "", "daemon", "stop")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	configText := fmt.Sprintf("version = 1\ndefault_profile = \"timed\"\n[profiles.timed]\nemail = %q\nlock_timeout = \"1s\"\n[profiles.steady]\nemail = %q\nlock_timeout = \"0s\"\n", manifest.Email, manifest.Email)
	if err := os.WriteFile(configPath, []byte(configText), 0600); err != nil {
		t.Fatal(err)
	}
	must("", "config valid\n", "--config", configPath, "config", "validate")
	must(manifest.Password+"\n", "", "--config", configPath, "login", "--server", endpoint.URL, "--password-stdin")
	must("", "STATE\tready\nVERSION\t"+version+"\nCONFIG\t"+configPath+"\n", "daemon", "status")
	verbose, err := run("", "--profile", "timed", "status", "--verbose")
	if err != nil || !strings.Contains(verbose, "SERVER\t"+endpoint.URL+"\n") || !strings.Contains(verbose, "EMAIL\t"+manifest.Email+"\n") || !strings.Contains(verbose, "LOCK_TIMEOUT\t1s\n") || strings.Contains(verbose, "LOCK_AT\t-\n") {
		t.Fatal("applied status", verbose, err)
	}
	must("", verbose, "--profile", "timed", "status", "--verbose")
	tuiTerminal("login", "--config", configPath, "--profile", "steady", "tui")

	// Policy changes require reload, apply per profile, and do not implicitly reveal details.
	skipConfig := configText + "insecure_skip_reprompt = true\n"
	if err := os.WriteFile(configPath, []byte(skipConfig), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run("", "--profile", "steady", "item", "read", protectedID, "--field", "password"); err == nil {
		t.Fatal("disk edit bypassed reprompt")
	}
	must("", "", "daemon", "reload")
	must("", "DEMO-reprompt-password!", "--profile", "steady", "item", "read", protectedID, "--field", "password")
	if output, err := run("", "--profile", "steady", "item", "read", protectedID); err != nil || strings.Contains(output, "DEMO-reprompt-password!") {
		t.Fatal("skip revealed secrets", err)
	}
	if _, err := run("", "--profile", "default", "item", "read", protectedID, "--field", "password"); err == nil {
		t.Fatal("skip crossed profile")
	}
	if err := os.WriteFile(configPath, []byte(configText), 0600); err != nil {
		t.Fatal(err)
	}
	must("", "", "daemon", "reload")
	if _, err := run("", "--profile", "steady", "item", "read", protectedID, "--field", "password"); err == nil {
		t.Fatal("skip remained after reload")
	}

	waitLocked := func() {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			output, err := run("", "--profile", "timed", "status")
			if err != nil {
				t.Fatal(err, output)
			}
			if output == "locked\n" {
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
		t.Fatal("idle timeout did not lock profile")
	}
	waitLocked()
	must("", "unlocked\n", "--profile", "steady", "status")
	offline.Store(true)
	before = requests.Load()
	terminal("success", "DEMO-only-password-1!", "--config", configPath, "item", "read", id, "--field", "password")
	if requests.Load() != before {
		t.Fatal("config/idle recovery contacted server")
	}
	// Changing config cannot rebind existing credentials.
	changed := strings.Replace(configText, manifest.Email, "wrong@example.invalid", 1)
	if err := os.WriteFile(configPath, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := run("", "daemon", "reload"); err == nil || !strings.Contains(output, "identity differs") {
		t.Fatal("configured identity mismatch was not blocked", err, output)
	}
	if err := os.WriteFile(configPath, []byte(configText), 0600); err != nil {
		t.Fatal(err)
	}
	must("", "", "daemon", "stop")
	// Restart loads the same explicitly selected config.
	must(manifest.Password+"\n", "", "--config", configPath, "--profile", "timed", "unlock", "--password-stdin")
	waitLocked()
	if requests.Load() != before {
		t.Fatal("restart unlock/status contacted server")
	}
	offline.Store(false)
	must(manifest.Password+"\n", "", "--config", configPath, "unlock", "--password-stdin")
	must("", "", "--config", configPath, "sync")
	waitLocked()
	// Direct edits do not affect live policy until a successful reload.
	must(manifest.Password+"\n", "", "--profile", "steady", "unlock", "--password-stdin")
	changed = strings.Replace(configText, `lock_timeout = "0s"`, `lock_timeout = "100ms"`, 1)
	if err := os.WriteFile(configPath, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond)
	must("", "PROFILE\tsteady\nSTATE\tunlocked\nSERVER\t"+endpoint.URL+"\nEMAIL\t"+manifest.Email+"\nLOCK_TIMEOUT\t0s\nLOCK_AT\t-\n", "--profile", "steady", "status", "--verbose")
	must("", "unlocked\n", "--profile", "steady", "status")
	must("", "", "daemon", "reload")
	must("", "locked\n", "--profile", "steady", "status")
	if err := os.WriteFile(configPath, []byte("version = ["), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run("", "daemon", "reload"); err == nil {
		t.Fatal("malformed reload accepted")
	}
	must("", "STATE\tready\nVERSION\t"+version+"\nCONFIG\t"+configPath+"\n", "--config", "does-not-exist.toml", "daemon", "status")
	must("", "PROFILE\tsteady\nSTATE\tlocked\nSERVER\t"+endpoint.URL+"\nEMAIL\t"+manifest.Email+"\nLOCK_TIMEOUT\t100ms\nLOCK_AT\t-\n", "--config", configPath, "--profile", "steady", "status", "--verbose")
	must(manifest.Password+"\n", "", "--config", configPath, "--profile", "timed", "unlock", "--password-stdin")
	waitLocked()
	must("", "", "--config", configPath, "--profile", "steady", "lock")
	must("", "", "--config", configPath, "--profile", "timed", "logout")
	must("", "", "--config", configPath, "--profile", "steady", "logout")
	if err := os.WriteFile(configPath, []byte(configText), 0600); err != nil {
		t.Fatal(err)
	}

	restrictedConfig := configText + fmt.Sprintf("\n[profiles.restricted]\nemail = %q\ninsecure_skip_reprompt = true\n", manifest.Organizations.RestrictedEmail)
	if err := os.WriteFile(configPath, []byte(restrictedConfig), 0600); err != nil {
		t.Fatal(err)
	}
	must("", "", "daemon", "reload")
	must(manifest.Password+"\n", "", "--profile", "restricted", "login", "--server", endpoint.URL, "--password-stdin")
	if output, err := run("", "--profile", "restricted", "item", "read", manifest.Organizations.Entries[0].Items[1].ID, "--field", "password"); err == nil || !strings.Contains(output, "organization policy") {
		t.Fatal("skip bypassed organization permissions", err, output)
	}

}
