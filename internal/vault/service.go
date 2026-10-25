// Package vault owns profile sessions and their lifetime, independently of transport.
package vault

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	daemon "github.com/134130/go-daemonkit"
	"github.com/134130/go-daemonkit/storage"
	"github.com/google/uuid"
)

var errLocked = &daemon.RPCError{Code: vaultapi.LockedCode, Message: "vault locked; run unlock or login"}

type profile struct {
	gate       chan struct{}
	epoch      uint64
	barriers   int
	cancel     context.CancelFunc
	session    *bitwarden.Session
	snapshot   *bitwarden.Snapshot
	credential string
}

// The gate owns session and persistent profile access; mu owns cancellation,
// generations and publication. Network and native I/O never hold mu.
type Service struct {
	Store    *storage.Store
	Secrets  Credentials
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
	var epoch uint64
	if p != nil {
		epoch = p.epoch
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
	if s.closed || s.profiles[name] != p || p.epoch != epoch || p.barriers != 0 || ctx.Err() != nil {
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
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	if saved.Deleting {
		return vaultapi.Empty{}, errors.New("logout incomplete; retry logout")
	}
	if saved.Cache != nil {
		if !client.Matches(saved.Cache) {
			return vaultapi.Empty{}, errors.New("profile identity differs; logout before changing account or server")
		}
		client, err = bitwarden.NewClient(saved.Cache.Server, saved.Cache.Email, saved.Cache.DeviceID)
		if err != nil {
			return vaultapi.Empty{}, err
		}
	}
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
	cache := session.Cache(snapshot)
	if saved.Cache != nil && saved.Cache.AccountID() != cache.AccountID() {
		return vaultapi.Empty{}, errors.New("profile account ID changed; logout first")
	}
	token := session.RefreshToken()
	defer clear(token)
	if err := s.saveLogin(ctx, saved, cache, token); err != nil {
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
	p.session, p.snapshot, p.credential = session, snapshot, saved.Credential
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
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	if saved.Cache == nil || saved.Deleting {
		return vaultapi.Empty{}, errLocked
	}
	if saved.Credential != p.credential {
		return vaultapi.Empty{}, errors.New("saved login changed; lock and unlock before syncing")
	}
	if err := s.cleanup(ctx, saved); err != nil {
		return vaultapi.Empty{}, err
	}
	token := p.session.RefreshToken()
	if len(token) == 0 {
		token, err = s.token(ctx, saved)
		if err != nil {
			return vaultapi.Empty{}, err
		}
		if err := p.session.SetRefreshToken(token); err != nil {
			clear(token)
			return vaultapi.Empty{}, err
		}
	}
	defer clear(token)
	snapshot, syncErr := p.session.SyncSnapshot(ctx)
	refreshed := p.session.RefreshToken()
	defer clear(refreshed)
	// Persist rotation even when the subsequent sync fails. Native calls may outlive cancellation.
	// The profile gate keeps logout from racing this write.
	writeCtx := ctx
	if !bytes.Equal(token, refreshed) {
		writeCtx = context.WithoutCancel(ctx)
	}
	// Also retry a previous credential write failure before reporting success.
	if err := s.saveToken(writeCtx, saved, refreshed); err != nil {
		return vaultapi.Empty{}, errors.Join(syncErr, err)
	}
	if syncErr != nil {
		return vaultapi.Empty{}, syncErr
	}
	saved.Cache = p.session.Cache(snapshot)
	if err := s.put(ctx, req.Profile, saved); err != nil {
		return vaultapi.Empty{}, err
	}
	if err := ctx.Err(); err != nil {
		return vaultapi.Empty{}, err
	}
	p.snapshot = snapshot
	return vaultapi.Empty{}, nil
}

func (s *Service) Unlock(ctx context.Context, req vaultapi.UnlockRequest) (vaultapi.Empty, error) {
	defer clear(req.Password)
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	defer done()
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	if saved.Deleting {
		return vaultapi.Empty{}, errors.New("logout incomplete; retry logout")
	}
	if saved.Cache == nil {
		return vaultapi.Empty{}, errors.New("no saved login; run login")
	}
	session, snapshot, err := bitwarden.Unlock(ctx, saved.Cache, req.Password)
	if err != nil {
		return vaultapi.Empty{}, err
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return vaultapi.Empty{}, err
	}
	if s.closed || p.barriers != 0 {
		return vaultapi.Empty{}, errLocked
	}
	if p.session != nil {
		p.session.Close()
	}
	p.session, p.snapshot, p.credential = session, snapshot, saved.Credential
	session = nil
	return vaultapi.Empty{}, nil
}

func (s *Service) Status(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.StatusResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return vaultapi.StatusResponse{}, err
	}
	defer done()
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return vaultapi.StatusResponse{}, err
	}
	state := "logged_out"
	if saved.Deleting {
		state = "logout_pending"
	} else if saved.Cache != nil {
		state = "locked"
		if p.session != nil {
			state = "unlocked"
		}
	}
	return vaultapi.StatusResponse{State: state}, nil
}

func destroy(p *profile) {
	p.gate <- struct{}{}
	defer func() { <-p.gate }()
	if p.session != nil {
		p.session.Close()
	}
	p.session, p.snapshot, p.credential = nil, nil, ""
}

func (s *Service) Lock(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	return vaultapi.Empty{}, s.clearProfile(ctx, req.Profile, false)
}
func (s *Service) Logout(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	return vaultapi.Empty{}, s.clearProfile(ctx, req.Profile, true)
}

func (s *Service) clearProfile(ctx context.Context, name string, logout bool) error {
	if err := validProfile(name); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errLocked
	}
	p := s.profiles[name]
	if p == nil {
		p = &profile{gate: make(chan struct{}, 1)}
		if s.profiles == nil {
			s.profiles = make(map[string]*profile)
		}
		s.profiles[name] = p
	}
	// Keep the gate across lock/logout; a replacement must not overtake persistent cleanup.
	p.epoch++
	p.barriers++
	if p.cancel != nil {
		p.cancel()
	}
	s.mu.Unlock()
	p.gate <- struct{}{}
	defer func() { s.mu.Lock(); p.barriers--; s.mu.Unlock(); <-p.gate }()
	if p.session != nil {
		p.session.Close()
	}
	p.session, p.snapshot, p.credential = nil, nil, ""
	if !logout {
		return nil
	}
	// Once logout is accepted, finish its durable tombstone even if the caller disconnects.
	ctx = context.WithoutCancel(ctx)
	saved, err := s.load(ctx, name)
	if err != nil {
		return err
	}
	if saved.Credential != "" {
		saved.Garbage = append(saved.Garbage, saved.Credential)
	}
	saved.Cache, saved.Credential, saved.Deleting = nil, "", true
	if err := s.put(ctx, name, saved); err != nil {
		return err
	}
	if err := s.cleanup(ctx, saved); err != nil {
		return err
	}
	return s.put(ctx, name, nil)
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
