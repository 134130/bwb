package vault

import (
	"context"
	"errors"
	"time"

	"github.com/134130/bwb/internal/config"
)

// Startup runs after daemonkit opens storage and before it publishes Ready.
func (s *Service) Startup(ctx context.Context) error {
	file := s.InitialConfig
	s.InitialConfig = nil
	return s.ApplyConfig(ctx, file)
}

// ApplyConfig validates the complete replacement before changing any live policy.
// configMu excludes profile operations for the validation/publication transaction.
func (s *Service) ApplyConfig(ctx context.Context, file *config.File) error {
	s.configMu.Lock()
	defer s.configMu.Unlock()
	return s.applyConfig(ctx, file)
}

// ReloadConfig serializes reading and publication so an older read cannot win.
func (s *Service) ReloadConfig(ctx context.Context, path string) error {
	s.configMu.Lock()
	defer s.configMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := config.Load(path, true)
	if err != nil {
		return err
	}
	return s.applyConfig(ctx, file)
}

func (s *Service) applyConfig(ctx context.Context, file *config.File) error {
	next := map[string]config.Profile{}
	if file != nil {
		for name, policy := range file.Profiles {
			if err := validProfile(name); err != nil {
				return err
			}
			normalized, err := config.Normalize(policy)
			if err != nil {
				return err
			}
			next[name] = normalized
		}
	}
	names := map[string]bool{}
	for name := range next {
		names[name] = true
	}
	for name := range s.policies {
		names[name] = true
	}
	for name := range names {
		saved, err := s.load(ctx, name)
		if err != nil {
			return err
		}
		if saved.Deleting {
			return errors.New("logout incomplete; retry logout")
		}
		if saved.Cache != nil {
			policy, ok := next[name]
			if !ok || saved.Cache.Email != policy.Email {
				return errors.New("configured identity differs or profile removed; logout before reload")
			}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed {
		return ErrStateChanged
	}
	s.policies = next
	for name, p := range s.profiles {
		p.clearReprompt()
		p.idleTimeout = 0
		if policy, ok := next[name]; ok {
			p.idleTimeout, _ = time.ParseDuration(policy.LockTimeout)
		}
		s.scheduleIdle(p)
	}
	return nil
}

// All idle fields are owned by mu; session/index destruction additionally holds the gate.
func (s *Service) stopIdle(p *profile) {
	p.idleVersion++
	if p.idleTimer != nil {
		p.idleTimer.Stop()
		p.idleTimer = nil
	}
	p.idleDeadline = time.Time{}
}

func (s *Service) scheduleIdle(p *profile) {
	s.stopIdle(p)
	if p.barriers != 0 || p.session == nil || p.idleTimeout == 0 {
		return
	}
	p.idleDeadline = p.lastActivity.Add(p.idleTimeout)
	version := p.idleVersion
	p.idleTimer = time.AfterFunc(max(time.Duration(0), time.Until(p.idleDeadline)), func() { s.expire(p, version) })
}

func (s *Service) activateIdle(p *profile, policy config.Profile) {
	p.idleTimeout = 0
	if policy.LockTimeout != "" {
		p.idleTimeout, _ = time.ParseDuration(policy.LockTimeout)
	}
	p.lastActivity = time.Now()
	s.scheduleIdle(p)
}

// invalidate cancels both in-flight work and queued admissions before waiting for the gate.
func (s *Service) invalidate(p *profile) {
	s.stopIdle(p)
	p.epoch++
	p.barriers++
	s.finishAuth(p)
	if p.cancel != nil {
		p.cancel()
	}
}

func clearMemory(p *profile) {
	if p.session != nil {
		p.session.Close()
	}
	p.clearSnapshotState()
	p.session, p.snapshot, p.credential = nil, nil, ""
}

func (s *Service) expire(p *profile, version uint64) {
	s.mu.Lock()
	if s.closed || p.idleVersion != version || p.idleDeadline.IsZero() || time.Now().Before(p.idleDeadline) {
		s.mu.Unlock()
		return
	}
	s.invalidate(p)
	s.mu.Unlock()
	destroy(p)
	s.mu.Lock()
	p.barriers--
	s.mu.Unlock()
}

func (s *Service) touch(ctx context.Context, p *profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed || p.barriers != 0 {
		return ErrStateChanged
	}
	// A delayed timer must not let a request extend an already expired session.
	if !p.idleDeadline.IsZero() && !time.Now().Before(p.idleDeadline) {
		return ErrStateChanged
	}
	p.lastActivity = time.Now()
	s.scheduleIdle(p)
	return nil
}
