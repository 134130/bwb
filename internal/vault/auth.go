package vault

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const AuthTimeout = 2 * time.Minute

type AuthRequest struct {
	Profile string `json:"profile"`
	Action  string `json:"action"`
	Epoch   uint64 `json:"epoch"`
	Version uint64 `json:"version"`
}
type AuthResponse struct {
	Ready  bool   `json:"ready"`
	ID     string `json:"id,omitempty"`
	Server string `json:"server,omitempty"`
	Email  string `json:"email,omitempty"`
}
type AuthEndRequest struct {
	Profile string `json:"profile"`
	ID      string `json:"id"`
}

type authAttempt struct {
	id, action string
	done       chan struct{}
	timer      *time.Timer
	claimed    bool
	watched    bool
	succeeded  bool
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
	return &AuthRequired{Action: action, Profile: name, Epoch: p.epoch, Version: p.authVersion}
}

// BeginAuth elects one prompt owner. Waiting RPCs are canceled on disconnection;
// an abandoned returned reservation expires, without retaining a password.
func (s *Service) BeginAuth(ctx context.Context, req AuthRequest) (AuthResponse, error) {
	if req.Action != "login" && req.Action != "unlock" {
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
		if saved.Cache == nil {
			err := s.authRequired(waitCtx, p, req.Profile, "login")
			done()
			return AuthResponse{}, err
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
		attempt := &authAttempt{id: uuid.NewString(), action: req.Action, done: make(chan struct{})}
		p.auth = attempt
		attempt.timer = time.AfterFunc(AuthTimeout, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if p.auth == attempt && !attempt.claimed {
				s.finishAuth(p)
			}
		})
		result := AuthResponse{ID: attempt.id, Server: saved.Cache.Server, Email: saved.Cache.Email}
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
	p.auth.timer.Stop()
	return nil
}
func (s *Service) endAuth(p *profile, id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.auth != nil && p.auth.id == id {
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
}

// WatchAuth holds the prompt owner's connection open; disconnection releases its reservation.
func (s *Service) WatchAuth(ctx context.Context, req AuthEndRequest) (AuthWatchResponse, error) {
	if err := validProfile(req.Profile); err != nil {
		return AuthWatchResponse{}, err
	}
	s.mu.Lock()
	p := s.profiles[req.Profile]
	if p != nil && p.previousAuth != nil && p.previousAuth.id == req.ID {
		result := AuthWatchResponse{Succeeded: p.previousAuth.succeeded}
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
		succeeded := attempt.succeeded
		s.mu.Unlock()
		return AuthWatchResponse{Succeeded: succeeded}, nil
	case <-ctx.Done():
		_, _ = s.EndAuth(context.WithoutCancel(ctx), req)
		return AuthWatchResponse{}, ctx.Err()
	}
}
