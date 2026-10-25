package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/134130/bwb/internal/vaultapi"
)

func TestVerboseStatusEscapesValuesAndPropagatesOutputErrors(t *testing.T) {
	var out bytes.Buffer
	result := vaultapi.StatusResponse{State: "logged_out", Email: "user\n@example.invalid", LockTimeout: "0s"}
	if err := writeStatus(&out, "default", result); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "\n") != 6 || !strings.Contains(out.String(), "SERVER\t-\n") || !strings.Contains(out.String(), "LOCK_AT\t-\n") {
		t.Fatal(out.String())
	}
	if err := writeStatus(brokenListWriter{}, "default", result); err == nil {
		t.Fatal("ignored write error")
	}
}
