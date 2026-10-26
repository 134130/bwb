package vault

import (
	"context"
	"errors"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
)

type TOTPRequest struct {
	Profile    string `json:"profile"`
	ID         string `json:"id"`
	Background bool   `json:"background,omitempty"`
}
type TOTPResponse = bitwarden.OTP

func (s *Service) TOTP(ctx context.Context, req TOTPRequest) (TOTPResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return TOTPResponse{}, err
	}
	defer done()
	if p.session == nil {
		return TOTPResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	if req.ID == "" || len(req.ID) > 4096 {
		return TOTPResponse{}, errors.New("invalid item selector")
	}
	if err := s.loadCatalog(ctx, p); err != nil {
		return TOTPResponse{}, err
	}
	entry, err := resolveItem(p.catalog["item"], req.ID)
	if err != nil {
		return TOTPResponse{}, err
	}
	seed, err := p.session.Read(p.snapshot, entry.ID, "login/totp")
	if err != nil {
		return TOTPResponse{}, err
	}
	result, err := bitwarden.GenerateTOTP(seed, time.Now())
	if err != nil {
		return TOTPResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return TOTPResponse{}, err
	}
	if !req.Background {
		if err := s.touch(ctx, p); err != nil {
			return TOTPResponse{}, err
		}
	}
	return result, nil
}
