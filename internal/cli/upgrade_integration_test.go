//go:build integration

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAutomaticDaemonUpgrade(t *testing.T) {
	server := os.Getenv("BWB_INTEGRATION_URL")
	if server == "" {
		t.Fatal("run through dev/integration/run.py")
	}
	var manifest struct {
		Email, Password string
		Items           map[string]string
	}
	data, err := os.ReadFile("../../testdata/integration/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	state, err := os.MkdirTemp("/tmp", "bwb-upgrade-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(state)
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(state, "config"))
	t.Setenv("BWB_CONFIG", "")
	t.Setenv("BWB_PROFILE", "")
	t.Setenv("BWB_INSTANCE", "upgrade")
	oldPath, newPath := filepath.Join(state, "old"), filepath.Join(state, "new")
	for binary, version := range map[string]string{oldPath: "v0.3.0", newPath: "v0.4.0"} {
		command := exec.Command("go", "build", "-ldflags", "-X github.com/134130/bwb/internal/daemon.Version="+version, "-o", binary, "../..")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build: %v %s", err, output)
		}
	}
	run := func(binary, password string, args ...string) (string, string, error) {
		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, args...)
		command.Stdin = strings.NewReader(password)
		var out, stderr bytes.Buffer
		command.Stdout, command.Stderr = &out, &stderr
		err := command.Run()
		return out.String(), stderr.String(), err
	}
	defer func() { _, _, _ = run(newPath, "", "daemon", "stop") }()
	active, other := filepath.Join(state, "active.toml"), filepath.Join(state, "other.toml")
	for _, path := range []string{active, other} {
		if err := os.WriteFile(path, []byte("version = 1\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, stderr, err := run(oldPath, manifest.Password+"\n", "--config", active, "--profile", "default", "login", "--server", server, "--email", manifest.Email, "--password-stdin"); err != nil {
		t.Fatal(err, stderr)
	}
	id := manifest.Items["20000000-0000-4000-8000-000000000001"]
	args := []string{"--config", other, "--profile", "default", "item", "read", id, "--field", "password"}
	out, stderr, err := run(newPath, "", append([]string{"--non-interactive"}, args...)...)
	if err != nil || out != "DEMO-only-password-1!" || strings.Count(stderr, "daemon v0.3.0 is older than CLI v0.4.0") != 1 {
		t.Fatal("noninteractive banner/output failed", err)
	}
	status, _, err := run(newPath, "", "daemon", "status")
	if err != nil || !strings.Contains(status, "VERSION\tv0.3.0\n") || !strings.Contains(status, "CONFIG\t"+active+"\n") {
		t.Fatal("noninteractive restarted daemon", status, err)
	}
	// A malformed active file must not destroy the unlocked daemon.
	if err := os.WriteFile(active, []byte("version = ["), 0600); err != nil {
		t.Fatal(err)
	}
	failed := exec.Command("../../.dev/venv/bin/python", append([]string{"../../dev/integration/upgrade.py", newPath}, args...)...)
	failed.Env = append(os.Environ(), "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_EXPECT=DEMO-only-password-1!", "BWB_TEST_FAILURE=1")
	if output, err := failed.CombinedOutput(); err != nil {
		t.Fatalf("preflight: %v %s", err, output)
	}
	out, _, err = run(oldPath, "", args...)
	if err != nil || out != "DEMO-only-password-1!" {
		t.Fatal("failed upgrade destroyed session", err)
	}
	if err := os.WriteFile(active, []byte("version = 1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("../../.dev/venv/bin/python", append([]string{"../../dev/integration/upgrade.py", newPath}, args...)...)
	command.Env = append(os.Environ(), "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_EXPECT=DEMO-only-password-1!")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("terminal upgrade: %v %s", err, output)
	}
	status, _, err = run(newPath, "", "daemon", "status")
	if err != nil || !strings.Contains(status, "VERSION\tv0.4.0\n") || !strings.Contains(status, "CONFIG\t"+active+"\n") {
		t.Fatal("upgrade lost version/config", status, err)
	}
	out, stderr, err = run(oldPath, "", args...)
	if err != nil || out != "DEMO-only-password-1!" || stderr != "" {
		t.Fatal("older CLI failed after upgrade", err)
	}
	status, _, err = run(oldPath, "", "daemon", "status")
	if err != nil || !strings.Contains(status, "VERSION\tv0.4.0\n") {
		t.Fatal("downgraded", status, err)
	}
	// Manual downgrade remains explicit; upgrade also supports the absent default config.
	if _, stderr, err := run(oldPath, "", "daemon", "update"); err != nil {
		t.Fatal(err, stderr)
	}
	command = exec.Command("../../.dev/venv/bin/python", "../../dev/integration/upgrade.py", newPath, "--profile", "default", "item", "read", id, "--field", "password")
	command.Env = append(os.Environ(), "BWB_TEST_PASSWORD="+manifest.Password, "BWB_TEST_EXPECT=DEMO-only-password-1!")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("default config upgrade: %v %s", err, output)
	}
	status, _, err = run(newPath, "", "daemon", "status")
	if err != nil || !strings.Contains(status, "VERSION\tv0.4.0\n") || !strings.Contains(status, "CONFIG\t"+filepath.Join(state, "config", "bwb", "config.toml")+"\n") {
		t.Fatal("default config not preserved", status, err)
	}

}
