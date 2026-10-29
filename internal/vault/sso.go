package vault

import (
	"context"
	"errors"
	"fmt"

	"github.com/134130/bwb/internal/bitwarden"
)

type LoginPlanRequest struct {
	Profile    string `json:"profile"`
	Server     string `json:"server"`
	Email      string `json:"email"`
	SSO        bool   `json:"sso"`
	Identifier string `json:"identifier"`
}

type LoginPlanResponse struct {
	AuthResponse
	Epoch     uint64 `json:"epoch"`
	Version   uint64 `json:"version"`
	Method    string `json:"method"`
	Discovery string `json:"discovery,omitempty"`
}

func (s *Service) PlanLogin(ctx context.Context, req LoginPlanRequest) (LoginPlanResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return LoginPlanResponse{}, err
	}
	defer done()
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return LoginPlanResponse{}, err
	}
	if saved.Deleting {
		return LoginPlanResponse{}, errors.New("logout incomplete; retry logout")
	}
	if err := bitwarden.ValidateSSOIdentifier(req.Identifier); err != nil {
		return LoginPlanResponse{}, err
	}
	client, err := s.loginClient(LoginRequest{Profile: req.Profile, Server: req.Server, Email: req.Email}, saved, true)
	if err != nil {
		return LoginPlanResponse{}, err
	}
	result := LoginPlanResponse{Method: "password"}
	result.Server, result.Email = client.Identity()
	result.Identifier = req.Identifier
	if req.SSO || saved.AuthMethod == "sso" {
		result.Method = "sso"
		if result.Identifier == "" && saved.Cache != nil {
			result.Identifier = saved.Cache.SSOIdentifier
		}
	} else {
		identifiers, err := client.DiscoverSSO(ctx)
		switch {
		case errors.Is(err, bitwarden.ErrSSODiscoveryUnsupported):
			result.Discovery = "unsupported"
		case err != nil:
			return LoginPlanResponse{}, err
		case len(identifiers) == 0:
			result.Discovery = "not_found"
		case len(identifiers) == 1:
			result.Method, result.Identifier, result.Discovery = "sso", identifiers[0], "matched"
		default:
			return LoginPlanResponse{}, errors.New("multiple SSO organizations found; choose one with login --sso --identifier")
		}
	}
	if saved.AuthMethod == "apikey" && result.Method != "sso" {
		return LoginPlanResponse{}, s.authRequired(ctx, p, req.Profile, "login")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.Err() != nil || s.closed || p.barriers != 0 {
		return LoginPlanResponse{}, ErrStateChanged
	}
	result.Epoch, result.Version = p.epoch, p.authVersion
	return result, nil
}

type ExchangeSSORequest struct {
	AuthEndRequest
	bitwarden.SSOCode
}

func (s *Service) ExchangeSSO(ctx context.Context, req ExchangeSSORequest) (AuthResponse, error) {
	defer clear(req.Code)
	defer clear(req.Verifier)
	if req.ID == "" {
		return AuthResponse{}, ErrStateChanged
	}
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return AuthResponse{}, err
	}
	defer done()
	if err := s.claimAuth(ctx, p, req.ID, "sso"); err != nil {
		return AuthResponse{}, err
	}
	s.mu.Lock()
	attempt := p.auth
	if attempt == nil || attempt.id != req.ID {
		s.mu.Unlock()
		return AuthResponse{}, ErrStateChanged
	}
	if attempt.pending != nil {
		s.mu.Unlock()
		s.endAuth(p, req.ID)
		return AuthResponse{}, errors.New("SSO code already exchanged")
	}
	client, identifier := attempt.client, attempt.identifier
	s.mu.Unlock()
	succeeded := false
	defer func() {
		if !succeeded {
			s.endAuth(p, req.ID)
		}
	}()
	if identifier != "" {
		if req.Identifier != "" && req.Identifier != identifier {
			return AuthResponse{}, errors.New("SSO organization differs from requested organization")
		}
		req.Identifier = identifier
	}
	pending, err := client.ExchangeSSO(ctx, req.SSOCode)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("exchange SSO code: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.Err() != nil || p.auth != attempt || p.barriers != 0 || s.closed {
		pending.Close()
		return AuthResponse{}, ErrStateChanged
	}
	attempt.pending, attempt.claimed = pending, false
	succeeded = true
	return AuthResponse{Email: pending.Email()}, nil
}

type FinishSSOResponse struct {
	RetryPassword bool `json:"retryPassword,omitempty"`
}

func (s *Service) FinishSSO(ctx context.Context, req UnlockRequest) (FinishSSOResponse, error) {
	defer clear(req.Password)
	if req.AuthID == "" {
		return FinishSSOResponse{}, ErrStateChanged
	}
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return FinishSSOResponse{}, fmt.Errorf("begin SSO unlock: %w", err)
	}
	defer done()
	if err := s.claimAuth(ctx, p, req.AuthID, "sso"); err != nil {
		return FinishSSOResponse{}, err
	}
	s.mu.Lock()
	attempt := p.auth
	if attempt == nil || attempt.id != req.AuthID || attempt.pending == nil {
		s.mu.Unlock()
		s.endAuth(p, req.AuthID)
		return FinishSSOResponse{}, ErrStateChanged
	}
	// Detach while deriving/syncing so cancellation never destroys a buffer in use.
	pending := attempt.pending
	attempt.pending = nil
	s.mu.Unlock()
	defer func() {
		if pending != nil {
			pending.Close()
			s.endAuth(p, req.AuthID)
		}
	}()
	session, err := pending.UnlockSSO(ctx, req.Password)
	if errors.Is(err, bitwarden.ErrCipher) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if ctx.Err() != nil || p.auth != attempt || s.closed || p.barriers != 0 {
			return FinishSSOResponse{}, ErrStateChanged
		}
		attempt.pending, attempt.claimed = pending, false
		pending = nil
		return FinishSSOResponse{RetryPassword: true}, nil
	}
	if err != nil {
		return FinishSSOResponse{}, fmt.Errorf("unlock SSO vault: %w", err)
	}
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		session.Close()
		return FinishSSOResponse{}, err
	}
	if _, err := s.loginClient(LoginRequest{Profile: req.Profile, Server: attempt.server, Email: pending.Email()}, saved, true); err != nil {
		session.Close()
		return FinishSSOResponse{}, err
	}
	_, err = s.installLogin(ctx, p, LoginRequest{Profile: req.Profile, AuthID: req.AuthID}, saved, session, nil)
	return FinishSSOResponse{}, err
}
