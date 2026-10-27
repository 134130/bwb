package tui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
)

func copyRoot(t *testing.T) *Root {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	r := &Root{ctx: ctx, page: "explorer", size: Size{120, 34}, explorer: newExplorer(), state: vaultapi.StatusResponse{State: "unlocked"}, epoch: 3, busy: true, workKind: "copy", cancel: cancel}
	r.explorer.Focus = "detail"
	r.explorer.Detail.Fields = []bitwarden.ItemField{{Key: "password"}, {Key: "username"}}
	return r
}
func TestCopyResultNeverRevealsValue(t *testing.T) {
	r := copyRoot(t)
	copied := ""
	r.writeClipboard = func(value string) error { copied = value; return nil }
	r.Update(resultMsg{Epoch: 3, Kind: "copy", Value: vaultapi.ReadResponse{Value: "secret"}})
	if copied != "secret" || r.message != "Copied" || secretValue(&r.explorer.Detail) != "" || r.explorer.Detail.Revealed {
		t.Fatal("copy revealed or failed")
	}
}
func TestNavigationAndLockDiscardLateCopy(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: tea.KeyDown}, {Code: '?', Text: "?"}, {Code: 'l', Mod: tea.ModCtrl}} {
		r := copyRoot(t)
		r.writeClipboard = func(string) error { t.Fatal("stale value copied"); return nil }
		r.Update(key)
		r.Update(resultMsg{Epoch: 3, Kind: "copy", Value: vaultapi.ReadResponse{Value: "secret"}})
	}
}
func TestDetailCopyRequestsSelectedField(t *testing.T) {
	d := Detail{Node: &Node{Item: &vaultapi.ListEntry{ID: "id"}}, Fields: testFields("secret")}
	cmd := d.Update(tea.KeyPressMsg{Code: 'y', Text: "y"}, Size{60, 20})
	request := cmd().(copyFieldMsg)
	if request.ID != "id" || request.Field != "password" {
		t.Fatal(request)
	}
}
