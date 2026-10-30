package cli

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCommandsAndProfileResolution(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("BWB_CONFIG", "")
	t.Setenv("BWB_PROFILE", "")
	path := filepath.Join(dir, "test.toml")
	data := `version = 1
default_profile = "work"
[profiles.work]
lock_timeout = "15m"
`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--config", path, "config", "path"}, path},
		{[]string{"--config", path, "config", "validate"}, "config valid"},
		{[]string{"--config", path, "profile", "list"}, "15m0s"},
		{[]string{"--config", path, "profile", "show", "work"}, "15m0s"},
	} {
		root := NewRoot()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs(tc.args)
		if err := root.Execute(); err != nil || !strings.Contains(out.String(), tc.want) {
			t.Fatal(tc.args, out.String(), err)
		}
	}
	// Noninteractive login without stdin must fail before starting the daemon.
	root := NewRoot()
	root.SetArgs([]string{"--config", path, "login", "--non-interactive"})
	root.SetIn(strings.NewReader(""))
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatal("noninteractive login prompted for password", err)
	}
	for _, args := range [][]string{

		{"--config", path, "profile", "show", "missing"},
	} {
		root := NewRoot()
		root.SetArgs(args)
		if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "profile") {
			t.Fatal("unknown profile accepted", err)
		}
	}
}

func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BWB_CONFIG", "")
	t.Setenv("BWB_PROFILE", "")
}

func TestExplicitProfileCommandsIgnoreMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.toml")
	if err := os.WriteFile(path, []byte("version = ["), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"lock", "logout", "status", "sync"} {
		root := NewRoot()
		command, _, err := root.Find([]string{name})
		if err != nil {
			t.Fatal(err)
		}
		called := false
		command.RunE = func(*cobra.Command, []string) error { called = true; return nil }
		root.SetArgs([]string{"--config", path, "--profile", "default", name})
		if err := root.Execute(); err != nil || !called {
			t.Fatal(name, err)
		}
	}
}

func TestTUIRejectsPipedInputBeforeStartingDaemon(t *testing.T) {
	isolateConfig(t)
	root := NewRoot()
	root.SetIn(strings.NewReader(""))
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "terminal input and output") {
		t.Fatal(err)
	}
}

func TestDaemonUpdateRejectsInvalidConfigBeforeStarting(t *testing.T) {
	isolateConfig(t)
	state := filepath.Join(t.TempDir(), "state")
	t.Setenv("XDG_STATE_HOME", state)
	path := filepath.Join(t.TempDir(), "config.toml")
	for _, tc := range []struct{ text, want string }{
		{"[profiles.default]\ninsecure_skip_reprompt = true\n", "add version = 1"},
		{"version = 1\n[profiles.default]\nemail = 'DO-NOT-ECHO@example.invalid'\n", "email is no longer a config field"},
		{"version = 1\n[profiles.default]\nlock_timeout = -1\n", "line 3"},
	} {
		if err := os.WriteFile(path, []byte(tc.text), 0600); err != nil {
			t.Fatal(err)
		}
		root := NewRoot()
		root.SetArgs([]string{"--config", path, "daemon", "update"})
		err := root.Execute()
		if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "DO-NOT-ECHO") {
			t.Fatalf("expected actionable config error, got %v", err)
		}
		if _, err := os.Stat(state); !os.IsNotExist(err) {
			t.Fatalf("invalid config reached daemon manager: %v", err)
		}
	}
}
