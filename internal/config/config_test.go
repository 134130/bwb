package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `version = 1
default_profile = "personal"
[profiles.personal]
server = "https://vault.bitwarden.com/"
email = "USER@example.com"
lock_timeout = "15m"
insecure_skip_reprompt = true
[profiles.work]
server = "http://localhost:8222"
email = "work@example.com"
`

func TestConfigValidationAndSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(sample), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := Load(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if file.Profiles["personal"].Email != "user@example.com" || file.Profiles["personal"].Server != "https://vault.bitwarden.com" || file.Profiles["personal"].LockTimeout != "15m0s" || file.Profiles["work"].LockTimeout != "0s" {
		t.Fatal("config normalization failed")
	}
	if !file.Profiles["personal"].InsecureSkipReprompt || file.Profiles["work"].InsecureSkipReprompt {
		t.Fatal("reprompt policy defaults")
	}
	t.Setenv("BWB_PROFILE", "")
	if name, err := Select("default", false, file); err != nil || name != "personal" {
		t.Fatal(name, err)
	}
	t.Setenv("BWB_PROFILE", "work")
	if name, err := Select("default", false, file); err != nil || name != "work" {
		t.Fatal(name, err)
	}
	if name, err := Select("personal", true, file); err != nil || name != "personal" {
		t.Fatal(name, err)
	}
	if _, err := Select("missing", true, file); err == nil {
		t.Fatal("unknown profile accepted")
	}
	if name, err := Select("legacy.Profile", true, nil); err != nil || name != "legacy.Profile" {
		t.Fatal("legacy profile rejected", err)
	}
	for _, bad := range []string{
		strings.Replace(sample, "insecure_skip_reprompt = true", `insecure_skip_reprompt = "true"`, 1),
		strings.Replace(sample, `email = "USER@example.com"`, `email = ""`, 1),
		strings.Replace(sample, `email = "USER@example.com"`, `email = " "`, 1),
		strings.Replace(sample, "version = 1", "version = 2", 1),
		strings.Replace(sample, `"15m"`, `"-1s"`, 1),
		strings.Replace(sample, `"15m"`, `"tomorrow"`, 1),
		strings.Replace(sample, `lock_timeout = "15m"`, `password = "DO-NOT-ECHO"`, 1),
		strings.Replace(sample, `default_profile = "personal"`, `default_profile = "missing"`, 1),
		strings.ReplaceAll(sample, "personal", "../bad"),
		sample + "token = 'DO-NOT-ECHO'\n", "version = [DO-NOT-ECHO",
	} {
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := Load(path, true)
		if err == nil || strings.Contains(err.Error(), "DO-NOT-ECHO") {
			t.Fatal("invalid config accepted or leaked", err)
		}
	}
}

func TestConfigPathsAndMissingFile(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("BWB_CONFIG", "")
	path, explicit, err := Path("")
	if err != nil || explicit || path != filepath.Join(base, "bwb", "config.toml") {
		t.Fatal(path, explicit, err)
	}
	if file, err := Load(path, false); err != nil || file != nil {
		t.Fatal("default missing file must preserve legacy use", err)
	}
	if _, err := Load(path, true); err == nil {
		t.Fatal("explicit missing config accepted")
	}
	t.Setenv("BWB_CONFIG", filepath.Join(base, "environment.toml"))
	if path, explicit, err := Path(""); err != nil || !explicit || filepath.Base(path) != "environment.toml" {
		t.Fatal(path, err)
	}
	if path, _, err := Path(filepath.Join(base, "flag.toml")); err != nil || filepath.Base(path) != "flag.toml" {
		t.Fatal(path, err)
	}
}
