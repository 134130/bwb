package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/134130/bwb/internal/vaultapi"
)

func TestListOutputEscapingGroupingAndJSON(t *testing.T) {
	collections := []vaultapi.NamedID{{ID: "first", Name: "Shared"}, {ID: "second", Name: "Shared"}}
	entries := []vaultapi.ListEntry{{ID: "id", Type: "login", Name: "long\nname\t\x1b[31m\u202e", Collections: collections}, {ID: "unassigned", Name: "한글"}}
	var out bytes.Buffer
	if err := writeList(&out, "item", "table", "", entries); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(out.String(), "\x1b\u202e") || !strings.Contains(out.String(), `long\nname\t\x1b`) || !strings.Contains(out.String(), "한글") {
		t.Fatal("terminal controls were not escaped", out.String())
	}
	out.Reset()
	if err := writeList(&out, "item", "table", "collection", entries); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"Shared [first]", "Shared [second]", "collection: —"} {
		if !strings.Contains(out.String(), label) {
			t.Fatal("groups merged or omitted", label)
		}
	}
	if strings.Count(out.String(), `long\nname`) != 2 {
		t.Fatal("multi-membership missing")
	}
	out.Reset()
	if err := writeList(&out, "item", "json", "collection", entries); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Version int
		Data    []vaultapi.ListEntry
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Version != 1 || len(result.Data) != 2 || result.Data[0].Name != entries[0].Name {
		t.Fatal("JSON grouping changed rows or raw names", err)
	}
	out.Reset()
	if err := writeList(&out, "item", "table", "", nil); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "\n") != 1 || !strings.HasPrefix(out.String(), "ID") {
		t.Fatal("empty table must have only a header")
	}
}

type brokenListWriter struct{}

func (brokenListWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestListOutputErrorsAndFlagValidation(t *testing.T) {
	isolateConfig(t)
	if err := writeList(brokenListWriter{}, "item", "table", "", nil); err == nil {
		t.Fatal("output error ignored")
	}
	for _, args := range [][]string{
		{"item", "list", "--format", "xml"}, {"item", "list", "--sort", "password"},
		{"item", "list", "--collection", "bad"}, {"item", "list", "--group-by", "type"},
		{"folder", "list", "--sort", "folder"},
	} {
		root := NewRoot()
		root.SetArgs(args)
		if err := root.Execute(); err == nil {
			t.Fatal("bad list flags accepted", args)
		}
	}
}
