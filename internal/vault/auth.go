package vault

import (
	"context"
	"errors"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/google/uuid"
)

const AuthTimeout = 2 * time.Minute
const SSOTimeout = 5 * time.Minute

type AuthRequest struct {
	Profile    string `json:"profile"`
	Action     string `json:"action"`
	Epoch      uint64 `json:"epoch"`
	Version    uint64 `json:"version"`
	Server     string `json:"server,omitempty"`
	Email      string `json:"email,omitempty"`
	Identifier string `json:"identifier,omitempty"`
}
type AuthResponse struct {
	Ready      bool   `json:"ready"`
	ID         string `json:"id,omitempty"`
	Server     string `json:"server,omitempty"`
	Email      string `json:"email,omitempty"`
	Identifier string `json:"identifier,omitempty"`
}
type AuthEndRequest struct {
	Profile string `json:"profile"`
	ID      string `json:"id"`
}

type authAttempt struct {
	id, action                string
	server, email, identifier string
	client                    *bitwarden.Client
	pending                   *bitwarden.Session
	done                      chan struct{}
	timer                     *time.Timer
	claimed                   bool
	watched                   bool
	succeeded                 bool
	failed                    bool
}

// authRequired is called under the profile gate. Infrastructure errors never become prompts.
func (s *Service) authRequired(ctx context.Context, p *profile, name, action string) error {
	saved, err := s.load(ctx, name)
	if err != nil {
		return err
	}
	if saved.Deleting {
		return errors.New("logout incomplete; retry logout")
	}
	if saved.Cache == nil {
		action = "login"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed || p.barriers != 0 {
		return ErrStateChanged
	}
	return &AuthRequired{Action: action, Method: saved.AuthMethod, Profile: name, Epoch: p.epoch, Version: p.authVersion}
}

// BeginAuth elects one prompt owner. Waiting RPCs are canceled on disconnection;
// an abandoned returned reservation expires, without retaining a password.
func (s *Service) BeginAuth(ctx context.Context, req AuthRequest) (AuthResponse, error) {
	if req.Action != "login" && req.Action != "unlock" && req.Action != "sso" {
		return AuthResponse{}, errors.New("invalid authentication action")
	}
	waited := false
	for {
		waitCtx, p, done, err := s.begin(ctx, req.Profile, true)
		if err != nil {
			return AuthResponse{}, err
		}
		saved, err := s.load(waitCtx, req.Profile)
		if err != nil {
			done()
			return AuthResponse{}, err
		}
		if saved.Deleting {
			done()
			return AuthResponse{}, errors.New("logout incomplete; retry logout")
		}
		if req.Action == "login" && saved.AuthMethod != "" {
			err := s.authRequired(waitCtx, p, req.Profile, "login")
			done()
			return AuthResponse{}, err
		}
		policy, configured := s.policies[req.Profile]
		if req.Action != "sso" && saved.Cache == nil && (!configured || req.Action != "login") {
			err := s.authRequired(waitCtx, p, req.Profile, "login")
			done()
			return AuthResponse{}, err
		}
		var client *bitwarden.Client
		if req.Action == "sso" {
			if err := bitwarden.ValidateSSOIdentifier(req.Identifier); err != nil {
				done()
				return AuthResponse{}, err
			}
			client, err = s.loginClient(LoginRequest{Profile: req.Profile, Server: req.Server, Email: req.Email}, saved, true)
			if err != nil {
				done()
				return AuthResponse{}, err
			}
			req.Server, req.Email = client.Identity()
		}
		s.mu.Lock()
		if waitCtx.Err() != nil || s.closed || p.barriers != 0 || p.epoch != req.Epoch {
			s.mu.Unlock()
			done()
			return AuthResponse{}, ErrStateChanged
		}
		if p.authVersion != req.Version {
			s.mu.Unlock()
			done()
			return AuthResponse{Ready: true}, nil
		}
		if waited {
			s.mu.Unlock()
			done()
			return AuthResponse{}, ErrStateChanged
		}
		if current := p.auth; current != nil {
			if req.Action == "sso" && (current.action != "sso" || current.server != req.Server || current.email != req.Email || current.identifier != req.Identifier) {
				s.mu.Unlock()
				done()
				return AuthResponse{}, errors.New("another authentication attempt is in progress for this profile")
			}
			ready := current.done
			s.mu.Unlock()
			done()
			select {
			case <-ready:
				waited = true
				continue
			case <-ctx.Done():
				return AuthResponse{}, ctx.Err()
			}
		}
		attempt := &authAttempt{id: uuid.NewString(), action: req.Action, done: make(chan struct{}), client: client,
			server: req.Server, email: req.Email, identifier: req.Identifier}
		p.auth = attempt
		timeout := AuthTimeout
		if req.Action == "sso" {
			timeout = SSOTimeout
		}
		attempt.timer = time.AfterFunc(timeout, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if p.auth == attempt && (!attempt.claimed || attempt.action == "sso") {
				if attempt.claimed && p.cancel != nil {
					p.cancel()
				}
				s.finishAuth(p)
			}
		})
		result := AuthResponse{ID: attempt.id, Email: policy.Email}
		if saved.Cache != nil {
			result.Server, result.Email = saved.Cache.Server, saved.Cache.Email
		}
		if req.Action == "sso" {
			result.Server, result.Email, result.Identifier = req.Server, req.Email, req.Identifier
		}
		s.mu.Unlock()
		done()
		return result, nil
	}
}

// finishAuth requires mu. Closing the reservation wakes all waiting clients.
func (s *Service) finishAuth(p *profile) {
	if p.auth == nil {
		return
	}
	if p.auth.timer != nil {
		p.auth.timer.Stop()
	}
	if p.auth.pending != nil {
		p.auth.pending.Close()
		p.auth.pending = nil
	}
	close(p.auth.done)
	p.previousAuth = p.auth
	p.auth = nil
}

func (s *Service) claimAuth(ctx context.Context, p *profile, id, action string) error {
	if id == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.auth == nil || p.auth.id != id || p.auth.action != action || p.auth.claimed || p.barriers != 0 {
		return ErrStateChanged
	}
	p.auth.claimed = true
	if action != "sso" {
		p.auth.timer.Stop()
	}
	return nil
}
func (s *Service) endAuth(p *profile, id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.auth != nil && p.auth.id == id {
		// The failing RPC carries the error; its watcher must not cancel that reply.
		p.auth.failed = true
		s.finishAuth(p)
	}
}
func (s *Service) EndAuth(ctx context.Context, req AuthEndRequest) (Empty, error) {
	if err := validProfile(req.Profile); err != nil {
		return Empty{}, err
	}
	if err := ctx.Err(); err != nil {
		return Empty{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.profiles[req.Profile]
	if p != nil && p.auth != nil && p.auth.id == req.ID {
		if p.auth.claimed && p.cancel != nil {
			p.cancel()
		}
		s.finishAuth(p)
	}
	return Empty{}, nil
}

type AuthWatchResponse struct {
	Succeeded bool `json:"succeeded"`
	Failed    bool `json:"failed,omitempty"`
}

// WatchAuth holds the prompt owner's connection open; disconnection releases its reservation.
func (s *Service) WatchAuth(ctx context.Context, req AuthEndRequest) (AuthWatchResponse, error) {
	if err := validProfile(req.Profile); err != nil {
		return AuthWatchResponse{}, err
	}
	s.mu.Lock()
	p := s.profiles[req.Profile]
	if p != nil && p.previousAuth != nil && p.previousAuth.id == req.ID {
		result := AuthWatchResponse{Succeeded: p.previousAuth.succeeded, Failed: p.previousAuth.failed}
		s.mu.Unlock()
		return result, nil
	}
	if p == nil || p.auth == nil || p.auth.id != req.ID || p.auth.watched {
		s.mu.Unlock()
		return AuthWatchResponse{}, ErrStateChanged
	}
	attempt := p.auth
	attempt.watched = true
	s.mu.Unlock()
	select {
	case <-attempt.done:
		s.mu.Lock()
		result := AuthWatchResponse{Succeeded: attempt.succeeded, Failed: attempt.failed}
		s.mu.Unlock()
		return result, nil
	case <-ctx.Done():
		_, _ = s.EndAuth(context.WithoutCancel(ctx), req)
		return AuthWatchResponse{}, ctx.Err()
	}
}
