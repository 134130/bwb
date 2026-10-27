package daemon

import (
	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/vaultapi"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpgradeVersionPolicy(t *testing.T) {
	for _, tc := range []struct {
		cli, daemon string
		want        bool
	}{
		{"v0.4.0", "v0.3.0", true}, {"v0.10.0", "v0.9.0", true},
		{"v0.3.0", "v0.4.0", false}, {"v0.3.0", "v0.3.0", false},
		{"dev", "v0.3.0", false}, {"v0.4.0", "dev", false},
		{"v0.4.0-rc.1", "v0.3.0", false}, {"v0.4.0", "v0.4.0-rc.1", false},
		{"v0.4.0+build", "v0.4.0", false}, {"bad\nversion", "v0.3.0", false},
	} {
		if got := newer(tc.cli, tc.daemon); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}

func TestLaunchEnvironmentPreservesParent(t *testing.T) {
	t.Setenv("BWB_CONFIG", "/caller/config.toml")
	t.Setenv("HTTPS_PROXY", "http://proxy.invalid")
	for _, path := range []string{"", "/active/config.toml"} {
		env := launchEnvironment(path)
		count := 0
		proxy := false
		for _, entry := range env {
			if strings.HasPrefix(entry, "BWB_CONFIG=") {
				count++
				if entry != "BWB_CONFIG="+path {
					t.Fatal(entry)
				}
			}
			if entry == "HTTPS_PROXY=http://proxy.invalid" {
				proxy = true
			}
		}
		if (path == "" && count != 0) || (path != "" && count != 1) || !proxy {
			t.Fatal(env)
		}
		if os.Getenv("BWB_CONFIG") != "/caller/config.toml" {
			t.Fatal("parent environment changed")
		}
	}
}

func TestUpgradePreservesOptionalConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BWB_CONFIG", "/caller/other.toml")
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	status := vaultapi.DaemonStatusResponse{ConfigPath: path, ConfigOptional: true}
	for _, exists := range []bool{false, true} {
		if exists {
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("version = 1\n"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		got, err := upgradeConfigPath(status)
		if err != nil || got != "" {
			t.Fatalf("optional config became required: %q %v", got, err)
		}
	}
	status.ConfigOptional = false
	if got, err := upgradeConfigPath(status); err != nil || got != path {
		t.Fatalf("explicit path lost: %q %v", got, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := upgradeConfigPath(status); err == nil {
		t.Fatal("missing explicit config accepted")
	}
}
