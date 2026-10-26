package cli

import (
	"bytes"
	"encoding/json"
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
