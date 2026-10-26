package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/vaultapi"
)

func TestOTPCopiesImmediatelyBeforeExpiryAndClearsOnClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := &Root{ctx: context.Background(), explorer: newExplorer(), otp: &OTPPage{ID: "item", Now: time.Now(), Code: vaultapi.TOTPResponse{Code: "123456", ExpiresAt: time.Now().Add(time.Second)}}}
		start := time.Now()
		copied := ""
		r.copyOTP(func(code string) error { copied = code; return nil })
		if copied != "123456" || time.Now() != start || r.otp.Message != "Copied" {
			t.Fatal("copy waited or failed")
		}
		r.copyOTP(func(string) error { return errors.New("unavailable") })
		if r.otp.Message != "Clipboard unavailable" {
			t.Fatal(r.otp.Message)
		}
		time.Sleep(time.Second)
		r.otp.Update(tickMsg(time.Now()), Size{})
		if strings.Contains(r.otp.View(Size{80, 20}), "123456") {
			t.Fatal("expired code displayed")
		}
		r.copyOTP(func(string) error { t.Fatal("expired code copied"); return nil })
		oldEpoch := r.epoch
		_, cmd := r.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		if r.otp != nil || cmd != nil {
			t.Fatal("TUI close should return to explorer")
		}
		r.Update(resultMsg{Epoch: oldEpoch, Kind: "otp", Value: vaultapi.TOTPResponse{Code: "late"}})
		if r.otp != nil {
			t.Fatal("late response reopened OTP")
		}
	})
}

func TestOTPLockPurgesCode(t *testing.T) {
	r := &Root{ctx: context.Background(), explorer: newExplorer(), otp: &OTPPage{Code: vaultapi.TOTPResponse{Code: "123456"}}, state: vaultapi.StatusResponse{State: "unlocked"}}
	r.Update(statusMsg{Epoch: r.epoch, Status: vaultapi.StatusResponse{State: "locked"}})
	if r.otp != nil {
		t.Fatal("lock retained code")
	}
}
