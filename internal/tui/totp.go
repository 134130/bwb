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
	profile, id, client, executor := r.profile, r.otp.ID, r.client, r.executor
	return r.operation("otp", func(ctx context.Context) (any, error) {
		req := vaultapi.TOTPRequest{Profile: profile, ID: id, Background: !authenticate}
		if authenticate {
			return executor.TOTP(ctx, req)
		}
		return client.Invoke(ctx, vaultapi.TOTP, req)
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

// Each inline display owns its refresh result; hiding or replacing the item drops that owner.
type inlineOTPMsg struct {
	page *OTPPage
	code vaultapi.TOTPResponse
	err  error
}

func (r *Root) refreshInlineOTP() tea.Cmd {
	d := &r.explorer.Detail
	if !d.Revealed || d.Node == nil || d.Node.Item == nil {
		return nil
	}
	hasOTP := false
	for _, f := range d.Fields {
		if f.Key == "totp" {
			hasOTP = true
		}
	}
	if !hasOTP {
		return nil
	}
	if d.OTP == nil {
		d.OTP = &OTPPage{ID: d.Node.Item.ID}
	}
	d.OTP.Now = time.Now()
	if d.otpPending || d.OTP.Stopped || d.OTP.Now.Before(d.OTP.Code.ExpiresAt) {
		return nil
	}
	d.otpPending = true
	page, client, profile, lifetime := d.OTP, r.client, r.profile, r.ctx
	id := page.ID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(lifetime, 3*time.Second)
		defer cancel()
		code, err := client.Invoke(ctx, vaultapi.TOTP, vaultapi.TOTPRequest{Profile: profile, ID: id, Background: true})
		return inlineOTPMsg{page, code, err}
	}
}
func (r *Root) writeCopy(value string) {
	if err := r.writeClipboard(value); err != nil {
		r.message = "Clipboard unavailable"
	} else {
		r.message = "Copied"
	}
}
func (r *Root) hideItem() {
	if r.busy && r.workKind == "reveal" {
		if r.cancel != nil {
			r.cancel()
		}
		r.epoch++
		r.busy = false
	}
	r.explorer.Detail.clearValue()
}
