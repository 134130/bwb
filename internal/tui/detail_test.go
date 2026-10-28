package tui

import (
	"context"
	"image/color"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/134130/bwb/internal/vaultclient"
	"github.com/charmbracelet/x/ansi"
)

func testFields(secret string) []bitwarden.ItemField {
	return []bitwarden.ItemField{{Key: "password", Label: "Password", Value: &secret, Sensitive: true, State: "visible"}}
}
func testItem(secret string) vaultapi.ItemResponse {
	return vaultapi.ItemResponse{ItemDetails: bitwarden.ItemDetails{Fields: testFields(secret)}}
}
func secretValue(d *Detail) string {
	for _, f := range d.Fields {
		if f.Sensitive && f.Value != nil {
			return *f.Value
		}
	}
	return ""
}
func TestItemRevealSurvivesFieldAndPaneNavigation(t *testing.T) {
	r := &Root{ctx: t.Context(), explorer: newExplorer(), size: Size{120, 34}, page: "explorer", state: vaultapi.StatusResponse{State: "unlocked"}, lastActivity: time.Now()}
	r.explorer.Tree.Roots = sampleTree()
	r.explorer.Tree.rebuild("", Size{40, 20})
	r.explorer.Tree.Index = 3
	r.explorer.Detail = Detail{Node: r.explorer.Tree.current()}
	item := testItem("secret")
	username := "public-user"
	item.Fields = append(item.Fields, bitwarden.ItemField{Key: "username", Label: "Username", Value: &username, State: "visible"})
	r.explorer.Detail.setItem(item, true)
	r.explorer.Focus = "detail"
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyDown}, {Code: tea.KeyTab}, {Code: tea.KeyTab}} {
		r.Update(key)
		if !r.explorer.Detail.Revealed || secretValue(&r.explorer.Detail) != "secret" {
			t.Fatal("navigation hid item")
		}
	}
	r.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	if r.explorer.Detail.Revealed || secretValue(&r.explorer.Detail) != "" {
		t.Fatal("hide retained sensitive fields")
	}
	if !strings.Contains(r.explorer.Detail.View(Size{70, 25}), username) {
		t.Fatal("hide removed public values")
	}
}
func TestSelectingDirectoryDiscardsPendingReveal(t *testing.T) {
	r := &Root{ctx: t.Context(), explorer: newExplorer(), epoch: 3, busy: true, workKind: "reveal"}
	r.explorer.Tree.Roots = sampleTree()
	r.explorer.Tree.rebuild("", Size{40, 20})
	r.describe()
	r.Update(resultMsg{Epoch: 3, Kind: "reveal", Value: testItem("late")})
	if secretValue(&r.explorer.Detail) != "" {
		t.Fatal("late item response published on directory")
	}
}
func TestHiddenExpandedFieldNeverReveals(t *testing.T) {
	d := Detail{Node: &Node{Name: "Item", Item: &vaultapi.ListEntry{ID: "one"}}}
	d.setItem(testItem("secret"), true)
	d.clearValue()
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter}, Size{60, 20})
	if !d.Expanded || strings.Contains(d.View(Size{60, 20}), "secret") {
		t.Fatal("expand disclosed a hidden field")
	}
}
func TestInlineOTPResultsCannotRestoreHiddenValues(t *testing.T) {
	page := &OTPPage{ID: "one"}
	r := &Root{ctx: t.Context(), explorer: newExplorer()}
	r.explorer.Detail.Revealed = true
	r.explorer.Detail.OTP = page
	r.hideItem()
	r.Update(inlineOTPMsg{page: page, code: vaultapi.TOTPResponse{Code: "123456", ExpiresAt: time.Now().Add(time.Second)}})
	if r.explorer.Detail.OTP != nil {
		t.Fatal("hidden item accepted late OTP")
	}
}
func TestThemeAndFramesStayBounded(t *testing.T) {
	r := &Root{ctx: context.Background(), explorer: newExplorer(), state: vaultapi.StatusResponse{State: "unlocked"}, message: "Copied"}
	for _, c := range []color.Color{color.Black, color.White} {
		r.Update(tea.BackgroundColorMsg{Color: c})
		if r.light != (c == color.White) {
			t.Fatal("incorrect background theme")
		}
		for _, size := range []Size{{1, 1}, {40, 12}, {80, 24}, {120, 34}} {
			r.Update(tea.WindowSizeMsg{Width: size.Width, Height: size.Height})
			before := r.explorer.Detail.Offset
			view := r.View().Content
			if view != r.View().Content || r.explorer.Detail.Offset != before {
				t.Fatal("View mutated model")
			}
			if len(strings.Split(view, "\n")) > size.Height {
				t.Fatal("height overflow")
			}
			for _, line := range strings.Split(view, "\n") {
				if ansi.StringWidth(line) > size.Width {
					t.Fatal("width overflow")
				}
			}
			if size.Width == 120 && (!strings.Contains(view, "Copied") || !strings.Contains(view, "Enter open")) {
				t.Fatal("notice replaced hints")
			}
		}
	}
}

func TestInlineOTPCopyAndExpiry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := &Root{ctx: t.Context(), page: "explorer", explorer: newExplorer(), lastActivity: time.Now(), state: vaultapi.StatusResponse{State: "unlocked"}}
		r.explorer.Tree.Roots = sampleTree()
		r.explorer.Tree.rebuild("", Size{40, 20})
		r.explorer.Tree.Index = 3
		node := r.explorer.Tree.current()
		d := &r.explorer.Detail
		*d = Detail{Node: node, Revealed: true, Fields: []bitwarden.ItemField{{Key: "totp", Label: "Authentication code", Sensitive: true}}, OTP: &OTPPage{ID: node.Item.ID, Code: vaultapi.TOTPResponse{Code: "123456", ExpiresAt: time.Now().Add(time.Second)}, Now: time.Now()}}
		r.explorer.Focus = "detail"
		copied := ""
		r.writeClipboard = func(value string) error { copied = value; return nil }
		started := time.Now()
		r.Update(copyFieldMsg{ID: node.Item.ID, Field: "totp"})
		if copied != "123456" || time.Now() != started {
			t.Fatal("copy delayed")
		}
		time.Sleep(time.Second)
		cmd := r.refreshInlineOTP()
		if cmd == nil || !d.otpPending || strings.Contains(d.fieldValue(d.Fields[0]), "123456") {
			t.Fatal("expired code retained")
		}
		if r.refreshInlineOTP() != nil {
			t.Fatal("duplicate refresh")
		}
		r.Update(inlineOTPMsg{page: d.OTP, code: vaultapi.TOTPResponse{Code: "654321", ExpiresAt: time.Now().Add(30 * time.Second)}})
		if !strings.Contains(d.fieldValue(d.Fields[0]), "654321") || d.otpPending {
			t.Fatal("fresh code not published")
		}
	})
}

func TestRestrictedItemNeverPublishesFields(t *testing.T) {
	d := Detail{Node: &Node{Name: "Restricted", Item: &vaultapi.ListEntry{ID: "one"}}}
	d.setItem(vaultapi.ItemResponse{ItemDetails: bitwarden.ItemDetails{Restriction: "reprompt is not supported"}}, true)
	if d.Revealed || len(d.Fields) != 0 || !strings.Contains(d.View(Size{80, 20}), "reprompt") {
		t.Fatal("restriction lost")
	}
}

func TestNarrowProfilesCanInspectAndScrollConfiguration(t *testing.T) {
	p := ProfilesPage{Rows: []profileRow{{Name: "work", Status: vaultapi.StatusResponse{Email: "user@example.invalid", Server: "https://example.invalid"}}}, Path: "/a/long/config/path.toml"}
	size := Size{40, 12}
	p.Update(tea.KeyPressMsg{Code: tea.KeyTab}, size)
	if !p.DetailFocus || !strings.Contains(p.View(size), "Applied configuration") {
		t.Fatal("config pane inaccessible")
	}
	p.Update(tea.KeyPressMsg{Code: tea.KeyPgDown}, size)
	if !strings.Contains(p.View(size), "Config:") {
		t.Fatal("config details not scrollable")
	}
	p.Update(tea.KeyPressMsg{Code: tea.KeyEscape}, size)
	if p.DetailFocus || !strings.Contains(p.View(size), "Profiles") {
		t.Fatal("cannot return to profiles")
	}
}

func TestRepromptModalCancelKeepsItemAndRejectsLateResult(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	r := &Root{ctx: t.Context(), explorer: newExplorer(), size: Size{120, 34}, state: vaultapi.StatusResponse{State: "unlocked"}, cancel: cancel, busy: true, epoch: 4}
	r.explorer.Tree.Roots = sampleTree()
	r.explorer.Tree.rebuild("", Size{40, 20})
	r.explorer.Tree.Index = 3
	node := r.explorer.Tree.current()
	r.explorer.Detail = Detail{Node: node, Reprompt: true}
	r.auth = &authDialog{Input: newInput(true), Prompt: promptMsg{Context: ctx, Request: vaultclient.AuthPrompt{Action: "reprompt"}}}
	r.auth.Input.SetValue("must-not-echo")
	r.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if r.auth != nil || r.busy || ctx.Err() == nil || r.explorer.Tree.current() != node {
		t.Fatal("cancel lost navigation or kept prompt")
	}
	r.Update(resultMsg{Epoch: 4, Kind: "reveal", Value: testItem("late-secret")})
	if secretValue(&r.explorer.Detail) != "" {
		t.Fatal("canceled reveal restored data")
	}
}
