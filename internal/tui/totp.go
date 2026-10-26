package tui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/134130/bwb/internal/vaultapi"
)

// OTPPage holds generated codes only; seeds never reach this screen.
type OTPPage struct {
	ID      string
	Code    vaultapi.TOTPResponse
	Now     time.Time
	Message string
	Stopped bool
}

func (p *OTPPage) Update(msg tea.Msg, _ Size) tea.Cmd {
	if t, ok := msg.(tickMsg); ok {
		p.Now = time.Time(t)
	}
	return nil
}
func (p *OTPPage) View(size Size) string {
	body := "Loading…"
	if p.Stopped {
		body = "Code unavailable · Enter to retry"
	}
	if p.Code.Code != "" && p.Now.Before(p.Code.ExpiresAt) {
		seconds := int((p.Code.ExpiresAt.Sub(p.Now) + time.Second - 1) / time.Second)
		body = fmt.Sprintf("%s\n%d seconds remaining", p.Code.Code, seconds)
	}
	return box("Authentication code\n\n"+body+"\n\n"+safe(p.Message)+"\ny copy · q close", size)
}
func (r *Root) openOTP(id string) tea.Cmd {
	r.explorer.Detail.clearValue()
	r.otp = &OTPPage{ID: id, Now: time.Now()}
	return r.refreshOTP(true)
}
func (r *Root) refreshOTP(authenticate bool) tea.Cmd {
	profile, id := r.profile, r.otp.ID
	return r.operation("otp", func(ctx context.Context) (any, error) {
		req := vaultapi.TOTPRequest{Profile: profile, ID: id, Background: !authenticate}
		if authenticate {
			return r.executor.TOTP(ctx, req)
		}
		return r.client.Invoke(ctx, vaultapi.TOTP, req)
	})
}
func (r *Root) closeOTP() {
	if r.cancel != nil {
		r.cancel()
	}
	r.epoch++
	r.busy = false
	r.otp = nil
}
func (r *Root) copyOTP(write func(string) error) {
	p := r.otp
	if p.Code.Code == "" || !time.Now().Before(p.Code.ExpiresAt) {
		p.Message = "Code expired; refreshing"
		return
	}
	// Copy the displayed code immediately; clipboard ownership is never renewed automatically.
	if err := write(p.Code.Code); err != nil {
		p.Message = "Clipboard unavailable"
	} else {
		p.Message = "Copied"
	}
}
