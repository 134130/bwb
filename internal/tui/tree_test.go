package tui

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/charmbracelet/x/ansi"
)

func sampleTree() []*Node {
	return []*Node{{Key: "org", Name: "Engineering", Children: []*Node{
		{Key: "dev", Name: "Development", Children: []*Node{{Key: "github", Name: "GitHub", Item: &vaultapi.ListEntry{ID: "one", Name: "GitHub"}}, {Key: "token", Name: "Deploy token", Item: &vaultapi.ListEntry{ID: "two"}}}},
		{Key: "shared", Name: "Shared", Children: []*Node{{Key: "github2", Name: "GitHub", Item: &vaultapi.ListEntry{ID: "one"}}}},
	}}}
}
func keys(tree *Tree) []string {
	out := []string{}
	for _, r := range tree.Rows {
		out = append(out, r.Node.Key)
	}
	return out
}
func TestFilteredTreeAndRestoredNavigation(t *testing.T) {
	tree := Tree{Roots: sampleTree()}
	size := Size{40, 5}
	tree.rebuild("", size)
	tree.Index = 3
	tree.Offset = 2
	original := tree.key()
	tree.rebuild("gh", size)
	if got := strings.Join(keys(&tree), ","); got != "org,dev,github,shared,github2" {
		t.Fatal(got)
	}
	tree.rebuild("dvlp", size)
	if got := strings.Join(keys(&tree), ","); got != "org,dev,token,github" {
		t.Fatal(got)
	}
	tree.rebuild("notfound", size)
	if len(tree.Rows) != 0 {
		t.Fatal(tree.Rows)
	}
	tree.rebuild("", size)
	if tree.key() != original {
		t.Fatal("selection not restored", tree.key(), original)
	}
	tree.Update(tea.KeyPressMsg{Code: tea.KeyLeft}, size)
	before := tree.key()
	tree.rebuild("gh", size)
	tree.rebuild("", size)
	if tree.key() != before {
		t.Fatal("collapsed state not restored")
	}
}
func TestTreeBuildKeepsMultiCollectionMembershipAndEmptyContainers(t *testing.T) {
	org := &vaultapi.NamedID{ID: "org", Name: "Engineering"}
	catalog := map[string][]vaultapi.ListEntry{
		"organization": {{ID: org.ID, Name: org.Name}},
		"collection":   {{ID: "a", Name: "Development", Organization: org}, {ID: "b", Name: "Shared", Organization: org}, {ID: "empty", Name: "Empty", Organization: org}},
		"item":         {{ID: "one", Name: "GitHub", Organization: org, Collections: []vaultapi.NamedID{{ID: "a"}, {ID: "b"}}}},
	}
	tree := Tree{Roots: buildTree(catalog)}
	tree.rebuild("", Size{60, 30})
	count := 0
	empty := false
	for _, row := range tree.Rows {
		if row.Node.Item != nil && row.Node.Item.ID == "one" {
			count++
		}
		if row.Node.Name == "Empty" {
			empty = true
		}
	}
	if count != 2 || !empty {
		t.Fatal(count, empty)
	}
}
func TestLayoutAndViewArePureAndBounded(t *testing.T) {
	e := newExplorer()
	e.Tree.Roots = sampleTree()
	for _, size := range []Size{{0, 0}, {1, 1}, {40, 12}, {100, 24}} {
		e.Tree.rebuild("", size)
		e.Update(LayoutMsg{}, size)
		before := keys(&e.Tree)
		index, offset := e.Tree.Index, e.Tree.Offset
		first := e.View(size)
		if first != e.View(size) || !reflect.DeepEqual(before, keys(&e.Tree)) || e.Tree.Index != index || e.Tree.Offset != offset {
			t.Fatal("View mutated state")
		}
		if size.Height > 0 && len(strings.Split(first, "\n")) > size.Height {
			t.Fatal("height overflow", size)
		}
		for _, line := range strings.Split(first, "\n") {
			if ansi.StringWidth(line) > size.Width {
				t.Fatal("width overflow", size, line)
			}
		}
	}
	e.Search.SetValue("gh")
	e.Focus = "search"
	e.Tree.rebuild("gh", Size{40, 12})
	before := e.Tree.key()
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter}, Size{40, 12})
	if e.Tree.key() != before || e.Focus != "tree" {
		t.Fatal("search enter navigated tree")
	}
}
func TestLockPurgesUIAndRejectsLateSecret(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		root := &Root{ctx: context.Background(), profile: "work", page: "explorer", size: Size{100, 24}, explorer: newExplorer(), state: vaultapi.StatusResponse{State: "unlocked"}, epoch: 7}
		root.explorer.Tree.Roots = sampleTree()
		root.explorer.Tree.rebuild("gh", Size{40, 20})
		root.explorer.Search.SetValue("gh")
		root.explorer.Detail = Detail{Value: "secret", Revealed: true, Fields: []bitwarden.FieldInfo{{Path: "login/password"}}}
		root.Update(statusMsg{Epoch: 7, Status: vaultapi.StatusResponse{State: "locked"}})
		root.Update(resultMsg{Epoch: 7, Kind: "read", Value: vaultapi.ReadResponse{Value: "late-secret"}})
		if root.explorer.Detail.Value != "" || len(root.explorer.Tree.Rows) != 0 || root.explorer.Search.Value() != "" || strings.Contains(root.View().Content, "secret") {
			t.Fatal("locked UI retained plaintext")
		}
	})
}
func TestTerminalControlsAreNotRendered(t *testing.T) {
	value := safe("safe\x1b[2J\x00\u202eevil")
	if strings.ContainsAny(value, "\x1b\x00\u202e") {
		t.Fatal(value)
	}
}

func TestAuthenticationInputOwnsFocusAndValue(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	r := &Root{ctx: ctx, size: Size{100, 30}, explorer: newExplorer()}
	r.Update(promptMsg{Context: ctx, Reply: make(chan []byte)})
	r.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if r.auth == nil || r.auth.Input.Value() != "a" {
		t.Fatal("dialog input not focused")
	}
	if strings.Contains(r.View().Content, "Password: a") {
		t.Fatal("password rendered")
	}
	r.purge()
	if r.auth != nil {
		t.Fatal("prompt retained")
	}
}

func TestDuplicateContainerNamesRetainIdentity(t *testing.T) {
	catalog := map[string][]vaultapi.ListEntry{"folder": {{ID: "a", Name: "Work"}, {ID: "b", Name: "Work"}}, "item": {{ID: "first", Name: "One", Folder: &vaultapi.NamedID{ID: "a"}}, {ID: "second", Name: "Two", Folder: &vaultapi.NamedID{ID: "b"}}}}
	roots := buildTree(catalog)
	if len(roots[0].Children) != 2 || roots[0].Children[0].Key == roots[0].Children[1].Key || len(roots[0].Children[0].Children) != 1 || len(roots[0].Children[1].Children) != 1 {
		t.Fatal("containers merged")
	}
}

var (
	_ Component = (*Tree)(nil)
	_ Component = (*Detail)(nil)
	_ Component = (*Input)(nil)
	_ Component = (*ExplorerPage)(nil)
	_ Component = (*ProfilesPage)(nil)
)

func TestLeavingDetailDiscardsPendingRead(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: '?', Text: "?"}} {
		ctx, cancel := context.WithCancel(t.Context())
		r := &Root{ctx: ctx, profile: "default", page: "explorer", state: vaultapi.StatusResponse{State: "unlocked"}, size: Size{100, 24}, explorer: newExplorer(), epoch: 3, busy: true, workKind: "read", cancel: cancel}
		r.explorer.Focus = "detail"
		r.Update(key)
		r.Update(resultMsg{Epoch: 3, Kind: "read", Value: vaultapi.ReadResponse{Value: "secret"}})
		if r.explorer.Detail.Value != "" || r.epoch == 3 {
			t.Fatal("stale reveal survived focus change")
		}
		cancel()
	}
}
