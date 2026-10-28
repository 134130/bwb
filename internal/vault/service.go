// Package vault owns profile sessions and their lifetime, independently of transport.
package vault

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
	"github.com/134130/go-daemonkit/storage"
	"github.com/google/uuid"
)

type profile struct {
	reprompts                  map[string]*repromptGrant
	repromptMu                 sync.Mutex
	idleTimeout                time.Duration
	idleTimer                  *time.Timer
	idleVersion                uint64
	idleDeadline, lastActivity time.Time
	catalog                    bitwarden.Catalog
	catalogID                  string
	gate                       chan struct{}
	epoch                      uint64
	barriers                   int
	cancel                     context.CancelFunc
	session                    *bitwarden.Session
	snapshot                   *bitwarden.Snapshot
	credential                 string
	auth                       *authAttempt
	previousAuth               *authAttempt
	authVersion                uint64
	needsLogin                 bool
}

// The gate owns session and persistent profile access; mu owns cancellation,
// generations and publication. Network and native I/O never hold mu.
type Service struct {
	InitialConfig *config.File
	configMu      sync.RWMutex
	policies      map[string]config.Profile
	Store         *storage.Store
	Secrets       Credentials
	mu            sync.Mutex
	profiles      map[string]*profile
	closed        bool
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
	s.configMu.RLock()
	admitted := false
	defer func() {
		if !admitted {
			s.configMu.RUnlock()
		}
	}()
	if err := validProfile(name); err != nil {
		return nil, nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, nil, nil, ErrStateChanged
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
		return nil, nil, nil, ErrStateChanged
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
		return nil, nil, nil, ErrStateChanged
	}
	// Enforce deadlines on admission even if the timer goroutine has not run yet.
	if !p.idleDeadline.IsZero() && !time.Now().Before(p.idleDeadline) {
		s.invalidate(p)
		s.mu.Unlock()
		clearMemory(p)
		s.mu.Lock()
		p.barriers--
		if s.closed || s.profiles[name] != p || p.barriers != 0 || ctx.Err() != nil {
			s.mu.Unlock()
			<-p.gate
			if err := ctx.Err(); err != nil {
				return nil, nil, nil, err
			}
			return nil, nil, nil, ErrStateChanged
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	s.mu.Unlock()
	admitted = true
	return ctx, p, func() {
		defer s.configMu.RUnlock()
		s.mu.Lock()
		p.cancel = nil
		s.mu.Unlock()
		cancel()
		<-p.gate
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (Empty, error) {
	return s.login(ctx, req, nil)
}

func (s *Service) LoginAPIKey(ctx context.Context, req LoginAPIKeyRequest) (Empty, error) {
	defer clear(req.ClientSecret)
	return s.login(ctx, req.LoginRequest, &req.APIKey)
}

func (s *Service) login(ctx context.Context, req LoginRequest, apiKey *bitwarden.APIKey) (Empty, error) {
	defer clear(req.Password)
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return Empty{}, err
	}
	defer done()
	if err := s.claimAuth(ctx, p, req.AuthID, "login"); err != nil {
		return Empty{}, err
	}
	defer s.endAuth(p, req.AuthID)
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return Empty{}, err
	}
	if saved.Deleting {
		return Empty{}, errors.New("logout incomplete; retry logout")
	}
	if saved.AuthMethod == "apikey" && apiKey == nil {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "login")
	}
	policy, configured := s.policies[req.Profile]
	if configured {
		if req.Server == "" {
			req.Server = policy.Server
		}
		if req.Email == "" {
			req.Email = policy.Email
		}
	}
	if apiKey != nil && req.Email == "" && saved.Cache != nil {
		req.Email = saved.Cache.Email
	}
	client, err := bitwarden.NewClient(req.Server, req.Email, uuid.NewString())
	if err != nil {
		return Empty{}, err
	}
	if configured && !client.Matches(&bitwarden.Cache{Server: policy.Server, Email: policy.Email}) {
		return Empty{}, errors.New("login identity differs from configured profile")
	}
	if saved.Cache != nil {
		if !client.Matches(saved.Cache) {
			return Empty{}, errors.New("profile identity differs; logout before changing account or server")
		}
		client, err = bitwarden.NewClient(saved.Cache.Server, saved.Cache.Email, saved.Cache.DeviceID)
		if err != nil {
			return Empty{}, err
		}
	}
	var session *bitwarden.Session
	if apiKey != nil {
		session, err = client.LoginAPIKey(ctx, req.Password, *apiKey)
	} else {
		session, err = client.Login(ctx, req.Password)
	}
	if err != nil {
		return Empty{}, err
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	snapshot, err := session.SyncSnapshot(ctx)
	if err != nil {
		return Empty{}, err
	}
	cache := session.Cache(snapshot)
	if saved.Cache != nil && saved.Cache.AccountID() != cache.AccountID() {
		return Empty{}, errors.New("profile account ID changed; logout first")
	}
	token := session.RefreshToken()
	defer clear(token)
	auth := credential{Token: token}
	if apiKey != nil {
		auth.ClientID, auth.ClientSecret = apiKey.ClientID, apiKey.ClientSecret
	}
	if err := s.saveLogin(ctx, saved, cache, auth); err != nil {
		return Empty{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Empty{}, err
	}
	if s.profiles[req.Profile] != p || s.closed {
		return Empty{}, ErrStateChanged
	}
	if p.session != nil {
		p.session.Close()
	}
	p.clearSnapshotState()
	p.session, p.snapshot, p.credential = session, snapshot, saved.Credential
	s.activateIdle(p, s.policies[req.Profile])
	p.authVersion++
	p.needsLogin = false
	if p.auth != nil && p.auth.id == req.AuthID {
		p.auth.succeeded = true
	}
	s.finishAuth(p)
	session = nil
	return Empty{}, nil
}

func (s *Service) Read(ctx context.Context, req ReadRequest) (ReadResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return ReadResponse{}, err
	}
	defer done()
	if p.session == nil {
		return ReadResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	snapshot := p.snapshot
	if req.Field != "name" {
		snapshot, err = s.itemSnapshot(ctx, p, req.Profile, req.ID, req.RepromptToken)
		if err != nil {
			return ReadResponse{}, err
		}
	}
	value, err := p.session.Read(snapshot, req.ID, req.Field)
	if err != nil {
		return ReadResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return ReadResponse{}, err
	}
	// shortcut: cap fields at 10 KB for worst-case JSON escaping; add chunked RPC when larger fields are needed.
	if len(value) > 10000 {
		return ReadResponse{}, errors.New("field exceeds RPC output limit (10000 bytes)")
	}
	if err := s.touch(ctx, p); err != nil {
		return ReadResponse{}, err
	}
	return ReadResponse{Value: value}, nil
}

func (s *Service) Sync(ctx context.Context, req ProfileRequest) (Empty, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return Empty{}, err
	}
	defer done()
	if p.session == nil {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return Empty{}, err
	}
	if saved.Deleting {
		return Empty{}, errors.New("logout incomplete; retry logout")
	}
	if saved.Cache == nil {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "login")
	}
	if saved.Credential != p.credential {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	if p.needsLogin {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "login")
	}
	if err := s.cleanup(ctx, saved); err != nil {
		return Empty{}, err
	}
	if !p.session.HasCredentials() {
		auth, err := s.loadCredential(ctx, saved)
		if err != nil {
			return Empty{}, err
		}
		if saved.AuthMethod == "apikey" {
			err = p.session.SetAPIKey(bitwarden.APIKey{ClientID: auth.ClientID, ClientSecret: auth.ClientSecret})
		} else {
			err = p.session.SetRefreshToken(auth.Token)
		}
		auth.clear()
		if err != nil {
			return Empty{}, err
		}
	}
	token := p.session.RefreshToken()
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
	if saved.AuthMethod == "" {
		if err := s.saveToken(writeCtx, saved, refreshed); err != nil {
			return Empty{}, errors.Join(syncErr, err)
		}
	}
	if syncErr != nil {
		if errors.Is(syncErr, bitwarden.ErrAuthentication) || errors.Is(syncErr, bitwarden.ErrKeyChanged) {
			p.needsLogin = true
			return Empty{}, s.authRequired(ctx, p, req.Profile, "login")
		}
		return Empty{}, syncErr
	}
	saved.Cache = p.session.Cache(snapshot)
	if err := s.put(ctx, req.Profile, saved); err != nil {
		return Empty{}, err
	}
	if err := ctx.Err(); err != nil {
		return Empty{}, err
	}
	p.clearSnapshotState()
	p.snapshot = snapshot
	return Empty{}, nil
}

func (s *Service) Unlock(ctx context.Context, req UnlockRequest) (Empty, error) {
	defer clear(req.Password)
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return Empty{}, err
	}
	defer done()
	if err := s.claimAuth(ctx, p, req.AuthID, "unlock"); err != nil {
		return Empty{}, err
	}
	defer s.endAuth(p, req.AuthID)
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return Empty{}, err
	}
	if saved.Deleting {
		return Empty{}, errors.New("logout incomplete; retry logout")
	}
	if saved.Cache == nil {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "login")
	}
	session, snapshot, err := bitwarden.Unlock(ctx, saved.Cache, req.Password)
	if err != nil {
		return Empty{}, err
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Empty{}, err
	}
	if s.closed || p.barriers != 0 {
		return Empty{}, ErrStateChanged
	}
	if p.session != nil {
		p.session.Close()
	}
	p.clearSnapshotState()
	p.session, p.snapshot, p.credential = session, snapshot, saved.Credential
	s.activateIdle(p, s.policies[req.Profile])
	p.authVersion++
	if p.auth != nil && p.auth.id == req.AuthID {
		p.auth.succeeded = true
	}
	s.finishAuth(p)
	session = nil
	return Empty{}, nil
}

func (s *Service) Status(ctx context.Context, req ProfileRequest) (StatusResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return StatusResponse{}, err
	}
	defer done()
	saved, err := s.load(ctx, req.Profile)
	if err != nil {
		return StatusResponse{}, err
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
	result := StatusResponse{State: state, LockTimeout: "0s"}
	if policy, ok := s.policies[req.Profile]; ok {
		result.Server, result.Email, result.LockTimeout = policy.Server, policy.Email, policy.LockTimeout
		result.InsecureSkipReprompt = policy.InsecureSkipReprompt
	} else if saved.Cache != nil {
		result.Server, result.Email = saved.Cache.Server, saved.Cache.Email
	}
	s.mu.Lock()
	if !p.idleDeadline.IsZero() {
		result.LockAt = p.idleDeadline.UTC().Format(time.RFC3339Nano)
	}
	s.mu.Unlock()
	return result, nil
}

func destroy(p *profile) {
	p.gate <- struct{}{}
	defer func() { <-p.gate }()
	clearMemory(p)
}

func (s *Service) Lock(ctx context.Context, req ProfileRequest) (Empty, error) {
	return Empty{}, s.clearProfile(ctx, req.Profile, false)
}
func (s *Service) Logout(ctx context.Context, req ProfileRequest) (Empty, error) {
	return Empty{}, s.clearProfile(ctx, req.Profile, true)
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
		return ErrStateChanged
	}
	p := s.profiles[name]
	if p == nil {
		p = &profile{gate: make(chan struct{}, 1)}
		if s.profiles == nil {
			s.profiles = make(map[string]*profile)
		}
		s.profiles[name] = p
	}
	// Keep the gate across lock/logout; a replacement cannot overtake cleanup.
	s.invalidate(p)
	s.mu.Unlock()
	if logout {
		s.configMu.RLock()
		defer s.configMu.RUnlock()
	}
	p.gate <- struct{}{}
	defer func() { s.mu.Lock(); p.barriers--; s.mu.Unlock(); <-p.gate }()
	clearMemory(p)
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
		s.stopIdle(p)
		s.finishAuth(p)
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
