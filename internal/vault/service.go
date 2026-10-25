// Package vault owns profile sessions and their lifetime, independently of transport.
package vault

import (
	"context"
	"errors"
	"sync"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	daemon "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
)

var errLocked = &daemon.RPCError{Code: vaultapi.LockedCode, Message: "vault locked; run login"}

type profile struct {
	gate     chan struct{}
	cancel   context.CancelFunc
	session  *bitwarden.Session
	snapshot *bitwarden.Snapshot
}

// Service's zero value is ready for use. The gate owns session access; mu owns
// membership, cancellation and publication. Network I/O never holds mu.
type Service struct {
	mu       sync.Mutex
	profiles map[string]*profile
	closed   bool
}

func validProfile(name string) error {
	if name == "" || len(name) > 64 || name == "." || name == ".." {
		return errors.New("invalid profile")
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return errors.New("invalid profile")
		}
	}
	return nil
}

func (s *Service) begin(ctx context.Context, name string, create bool) (context.Context, *profile, func(), error) {
	if err := validProfile(name); err != nil {
		return nil, nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, nil, nil, errLocked
	}
	p := s.profiles[name]
	if p == nil && create {
		p = &profile{gate: make(chan struct{}, 1)}
		if s.profiles == nil {
			s.profiles = make(map[string]*profile)
		}
		s.profiles[name] = p
	}
	s.mu.Unlock()
	if p == nil {
		return nil, nil, nil, errLocked
	}
	select {
	case p.gate <- struct{}{}:
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	}
	s.mu.Lock()
	if s.closed || s.profiles[name] != p || ctx.Err() != nil {
		s.mu.Unlock()
		<-p.gate
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		return nil, nil, nil, errLocked
	}
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	s.mu.Unlock()
	return ctx, p, func() {
		s.mu.Lock()
		p.cancel = nil
		s.mu.Unlock()
		cancel()
		<-p.gate
	}, nil
}

func (s *Service) Login(ctx context.Context, req vaultapi.LoginRequest) (vaultapi.Empty, error) {
	defer clear(req.Password)
	client, err := bitwarden.NewClient(req.Server, req.Email, uuid.NewString())
	if err != nil {
		return vaultapi.Empty{}, err
	}
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	defer done()
	session, err := client.Login(ctx, req.Password)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	snapshot, err := session.SyncSnapshot(ctx)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return vaultapi.Empty{}, err
	}
	if s.profiles[req.Profile] != p || s.closed {
		return vaultapi.Empty{}, errLocked
	}
	if p.session != nil {
		p.session.Close()
	}
	p.session, p.snapshot = session, snapshot
	session = nil
	return vaultapi.Empty{}, nil
}

func (s *Service) Read(ctx context.Context, req vaultapi.ReadRequest) (vaultapi.ReadResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, false)
	if err != nil {
		return vaultapi.ReadResponse{}, err
	}
	defer done()
	if p.session == nil {
		return vaultapi.ReadResponse{}, errLocked
	}
	value, err := p.session.Read(p.snapshot, req.ID, req.Field)
	if err != nil {
		return vaultapi.ReadResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return vaultapi.ReadResponse{}, err
	}
	// shortcut: cap fields at 10 KB for worst-case JSON escaping; add chunked RPC when larger fields are needed.
	if len(value) > 10000 {
		return vaultapi.ReadResponse{}, errors.New("field exceeds RPC output limit (10000 bytes)")
	}
	return vaultapi.ReadResponse{Value: value}, nil
}

func (s *Service) Sync(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, false)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	defer done()
	if p.session == nil {
		return vaultapi.Empty{}, errLocked
	}
	snapshot, err := p.session.SyncSnapshot(ctx)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	if err := ctx.Err(); err != nil {
		return vaultapi.Empty{}, err
	}
	p.snapshot = snapshot
	return vaultapi.Empty{}, nil
}

func (s *Service) Status(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.StatusResponse, error) {
	if err := validProfile(req.Profile); err != nil {
		return vaultapi.StatusResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return vaultapi.StatusResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.profiles[req.Profile]
	return vaultapi.StatusResponse{Unlocked: p != nil && p.session != nil}, nil
}

func destroy(p *profile) {
	p.gate <- struct{}{}
	defer func() { <-p.gate }()
	if p.session != nil {
		p.session.Close()
	}
	p.session, p.snapshot = nil, nil
}

func (s *Service) Lock(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	if err := validProfile(req.Profile); err != nil {
		return vaultapi.Empty{}, err
	}
	if err := ctx.Err(); err != nil {
		return vaultapi.Empty{}, err
	}
	s.mu.Lock()
	p := s.profiles[req.Profile]
	delete(s.profiles, req.Profile)
	if p != nil && p.cancel != nil {
		p.cancel()
	}
	s.mu.Unlock()
	// Once accepted, finish destruction even if the caller disconnects.
	if p != nil {
		destroy(p)
	}
	return vaultapi.Empty{}, nil
}

func (s *Service) Shutdown(context.Context) error {
	s.mu.Lock()
	s.closed = true
	profiles := s.profiles
	s.profiles = nil
	for _, p := range profiles {
		if p.cancel != nil {
			p.cancel()
		}
	}
	s.mu.Unlock()
	for _, p := range profiles {
		destroy(p)
	}
	return nil
}
