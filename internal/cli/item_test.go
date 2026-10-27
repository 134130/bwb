package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
)

func TestItemOutputAndRawField(t *testing.T) {
	result := vaultapi.ItemResponse{Item: vaultapi.ListEntry{ID: "id", Name: "Example\x1b[31m", Type: "login"}, ItemDetails: bitwarden.ItemDetails{Fields: []bitwarden.ItemField{
		{Key: "username", Label: "Username", State: "visible", Value: new("user@example.invalid")},
		{Key: "password", Label: "Password", Sensitive: true, State: "hidden"},
		{Key: "totp-secret", Label: "TOTP seed", Sensitive: true, State: "hidden"},
		{Key: "custom/token", Label: "token", State: "empty", Value: new("")},
	}}}
	var out bytes.Buffer
	if err := writeItem(&out, "text", "work", result, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Login credentials", "[hidden]", "[empty]", "--field 'username'", "--profile 'work'", "item totp id", "--reveal"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal("missing discovery information", want, out.String())
		}
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Fatal("terminal escape leaked")
	}
	out.Reset()
	if err := writeItem(&out, "json", "work", result, false); err != nil {
		t.Fatal(err)
	}
	var decoded struct{ Data vaultapi.ItemResponse }
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil || decoded.Data.Fields[1].Value != nil || decoded.Data.Fields[1].State != "hidden" {
		t.Fatal("JSON lost masked state", err)
	}
	result.Fields = []bitwarden.ItemField{{Key: "notes", State: "visible", Value: new("line 1\nline 2\n")}}
	out.Reset()
	if err := writeItem(&out, "text", "work", result, true); err != nil || out.String() != "line 1\nline 2\n" {
		t.Fatal("raw value changed", err)
	}
	for _, format := range []string{"text", "json"} {
		if err := writeItem(brokenListWriter{}, format, "work", result, false); err == nil {
			t.Fatal("output failure ignored")
		}
	}
	if got := shellQuote("a'$(x)"); got != "'a'\\''$(x)'" {
		t.Fatal(got)
	}
}

func TestCopyItemDoesNotPrintValue(t *testing.T) {
	var out bytes.Buffer
	result := vaultapi.ItemResponse{ItemDetails: bitwarden.ItemDetails{Fields: []bitwarden.ItemField{{Value: new("secret")}}}}
	copied := ""
	if err := copyItem(t.Context(), &out, result, func(value string) error { copied = value; return nil }); err != nil || copied != "secret" || out.String() != "Copied\n" {
		t.Fatal("copy failed", err)
	}
	out.Reset()
	if err := copyItem(t.Context(), &out, result, func(string) error { return errors.New("secret") }); err == nil || strings.Contains(err.Error(), "secret") || out.Len() != 0 {
		t.Fatal("clipboard error leaked value")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := copyItem(ctx, &out, result, func(string) error { t.Fatal("copied after cancel"); return nil }); err == nil {
		t.Fatal("missing cancellation")
	}
}
func TestCopyFlagsRejectBeforeConnecting(t *testing.T) {
	for _, args := range [][]string{
		{"item", "read", "x", "--copy"},
		{"item", "read", "x", "--field", "password", "--copy", "--format", "json"},
		{"item", "read", "x", "--field", "password", "--copy", "--reveal"},
	} {
		cmd := NewRoot()
		cmd.SetArgs(append([]string{"--profile", "default"}, args...))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--copy") {
			t.Fatal("invalid copy flags not rejected", err)
		}
	}
}
