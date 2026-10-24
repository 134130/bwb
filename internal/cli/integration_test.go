//go:build integration

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestReadAgainstSnapshot(t *testing.T) {
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
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ fixtureID, path, want string }{
		{"20000000-0000-4000-8000-000000000001", "login/password", "DEMO-only-password-1!"},
		{"20000000-0000-4000-8000-000000000005", "login/password", ""},
		{"20000000-0000-4000-8000-000000000006", "notes", "DEMO-0001\nDEMO-0002\n"},
	} {
		t.Run(tc.path+tc.fixtureID, func(t *testing.T) {
			root := NewRoot()
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetIn(strings.NewReader(manifest.Password + "\n"))
			root.SetArgs([]string{"read", manifest.Items[tc.fixtureID], tc.path, "--server", server, "--email", manifest.Email, "--password-stdin"})
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := root.ExecuteContext(ctx); err != nil {
				t.Fatal(err)
			}
			if stdout.String() != tc.want || stderr.Len() != 0 {
				t.Fatal("CLI output mismatch")
			}
		})
	}
}
