package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

func TestReconnectFailureCanRetryWithoutRestoringSecrets(t *testing.T) {
	calls := 0
	r := &Root{ctx: context.Background(), explorer: newExplorer(), size: Size{120, 34},
		connect: func(context.Context) (*kit.Client, error) { calls++; return nil, errors.New("connection failed") },
		state:   vaultapi.StatusResponse{State: "unlocked"}}
	r.explorer.Tree.Roots = sampleTree()
	r.explorer.Tree.rebuild("gh", Size{60, 30})
	r.explorer.Search.SetValue("gh")
	r.explorer.Detail.Value = "secret"
	oldEpoch := r.epoch
	key := r.explorer.Tree.key()
	r.Update(statusMsg{Epoch: oldEpoch, Err: errors.New("closed")})
	r.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if r.explorer.Tree.savedKey != key {
		t.Fatal("cancel discarded navigation")
	}
	if r.state.State != "disconnected" || r.explorer.Detail.Value != "" || r.explorer.Search.Value() != "" || len(r.explorer.Tree.Roots) != 0 {
		t.Fatal("disconnect did not purge")
	}
	for range 2 {
		_, cmd := r.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if cmd == nil || !r.busy {
			t.Fatal("Enter did not reconnect")
		}
		_, duplicate := r.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if duplicate != nil {
			t.Fatal("duplicate connection attempt")
		}
		r.Update(cmd())
		if r.busy || r.state.State != "disconnected" || !strings.Contains(r.message, "connection failed") {
			t.Fatal("failure not retryable", r.message)
		}
	}
	r.Update(resultMsg{Epoch: oldEpoch, Kind: "read", Value: vaultapi.ReadResponse{Value: "late secret"}})
	if calls != 2 || r.explorer.Detail.Value != "" {
		t.Fatal("retry count or late response")
	}
}

func TestReconnectCancellationRejectsLateResult(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := &Root{ctx: context.Background(), explorer: newExplorer(), state: vaultapi.StatusResponse{State: "disconnected"},
			connect: func(ctx context.Context) (*kit.Client, error) { <-ctx.Done(); return nil, ctx.Err() }}
		_, cmd := r.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		done := make(chan tea.Msg, 1)
		go func() { done <- cmd() }()
		synctest.Wait()
		r.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		synctest.Wait()
		r.Update(<-done)
		if r.busy || r.state.State != "disconnected" || r.message != "Daemon disconnected · Enter to reconnect" {
			t.Fatal("cancel lost disconnected state")
		}
	})
}
