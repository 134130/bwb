package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestAPIKeyEnvironmentAndNoninteractiveValidation(t *testing.T) {
	isolateConfig(t)
	const clientID = "user.10000000-0000-4000-8000-000000000001"
	for _, test := range []struct {
		id, secret string
		valid      bool
	}{
		{clientID, "test-secret", true},
		{"", "test-secret", false},
		{clientID, "", false},
		{"organization.10000000-0000-4000-8000-000000000001", "test-secret", false},
		{clientID, strings.Repeat("x", 4097), false},
	} {
		t.Setenv("BWB_CLIENTID", test.id)
		t.Setenv("BWB_CLIENTSECRET", test.secret)
		cmd := NewRoot()
		input := strings.NewReader("master password\n")
		cmd.SetIn(input)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		key, err := commandAPIKey(cmd, true, authPrompt{})
		if (err == nil) != test.valid {
			t.Fatal("unexpected API key validation result", err)
		}
		if test.valid && (key.ClientID != test.id || string(key.ClientSecret) != test.secret) {
			t.Fatal("changed API credentials")
		}
		clear(key.ClientSecret)
		if input.Len() != len("master password\n") || out.Len() != 0 {
			t.Fatal("API key input consumed stdin or wrote output")
		}
		if _, ok := os.LookupEnv("BWB_CLIENTSECRET"); ok {
			t.Fatal("secret left for daemon inheritance")
		}
		if _, ok := os.LookupEnv("BWB_CLIENTID"); ok {
			t.Fatal("client ID left for daemon inheritance")
		}
	}
}

func TestAPIKeyFlagDoesNotFallBackToPassword(t *testing.T) {
	isolateConfig(t)
	t.Setenv("BWB_CLIENTID", "")
	t.Setenv("BWB_CLIENTSECRET", "")
	cmd := NewRoot()
	input := strings.NewReader("master password\n")
	cmd.SetIn(input)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"login", "--apikey", "--password-stdin", "--non-interactive"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "BWB_CLIENTSECRET") {
		t.Fatal("missing API key fell back to password", err)
	}
	if input.Len() != len("master password\n") {
		t.Fatal("missing API key consumed master password")
	}
}

func TestOtherCommandsDiscardAPIKeyEnvironment(t *testing.T) {
	isolateConfig(t)
	for _, args := range [][]string{{"config", "path"}, {"login", "--non-interactive"}} {
		t.Setenv("BWB_CLIENTID", "invalid-unused-id")
		t.Setenv("BWB_CLIENTSECRET", "unused-secret")
		cmd := NewRoot()
		cmd.SetIn(strings.NewReader("\n"))
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		err := cmd.Execute()
		if args[0] == "login" {
			if err == nil || !strings.Contains(err.Error(), "password") {
				t.Fatal("API key environment selected API login", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if os.Getenv("BWB_CLIENTID") != "" || os.Getenv("BWB_CLIENTSECRET") != "" {
			t.Fatal("unused credentials could reach daemon")
		}
	}
}
