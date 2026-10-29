package vault

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
)

func idleProfile(t *testing.T, s *Service, name, timeout string) *profile {
	t.Helper()
	_, p, done, err := s.begin(context.Background(), name, true)
	if err != nil {
		t.Fatal(err)
	}
	p.session = &bitwarden.Session{}
	p.catalog = bitwarden.Catalog{"item": {{Name: "sensitive name"}}}
	s.mu.Lock()
	s.activateIdle(p, config.Profile{LockTimeout: timeout})
	s.mu.Unlock()
	done()
	return p
}

func TestIdleActivityAndProfileIsolation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		defer s.Shutdown(context.Background())
		a := idleProfile(t, s, "a", "10s")
		b := idleProfile(t, s, "b", "10s")
		time.Sleep(6 * time.Second)
		ctx, p, done, err := s.begin(context.Background(), "a", false)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.touch(ctx, p); err != nil {
			t.Fatal(err)
		}
		done()
		// A policy reload, status or sync does not change lastActivity.
		b.gate <- struct{}{}
		s.mu.Lock()
		s.scheduleIdle(b)
		s.mu.Unlock()
		<-b.gate
		time.Sleep(4 * time.Second)
		synctest.Wait()
		if inspectIdle(s, b).session || inspectIdle(s, b).catalog || !inspectIdle(s, a).session {
			t.Fatal("idle expiry or profile isolation failed")
		}
		time.Sleep(6 * time.Second)
		synctest.Wait()
		if inspectIdle(s, a).session || inspectIdle(s, a).catalog {
			t.Fatal("successful read did not renew exact deadline")
		}
	})
}

func TestIdleCancelsActiveWorkAndRejectsStaleTimer(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		defer s.Shutdown(context.Background())
		p := idleProfile(t, s, "default", "10s")
		oldVersion := inspectIdle(s, p).version
		time.Sleep(5 * time.Second)
		ctx, _, done, err := s.begin(context.Background(), "default", false)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.touch(ctx, p); err != nil {
			t.Fatal(err)
		}
		done()
		s.expire(p, oldVersion)
		if !inspectIdle(s, p).session {
			t.Fatal("stale timer destroyed renewed session")
		}
		ctx, _, done, err = s.begin(context.Background(), "default", false)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatal("expiry did not cancel active operation")
		}
		if err := s.touch(ctx, p); err == nil {
			t.Fatal("expired work renewed session")
		}
		done()
		synctest.Wait()
		if inspectIdle(s, p).session || inspectIdle(s, p).timer || inspectIdle(s, p).barriers != 0 {
			t.Fatal("expiry did not finish cleanup")
		}
	})
}

func TestIdleDisabledManualLockAndShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		p := idleProfile(t, s, "default", "0s")
		time.Sleep(time.Hour)
		if !inspectIdle(s, p).session || inspectIdle(s, p).timer {
			t.Fatal("zero timeout expired")
		}
		if _, err := s.Lock(context.Background(), ProfileRequest{Profile: "default"}); err != nil {
			t.Fatal(err)
		}
		p = idleProfile(t, s, "default", "1s")
		if err := s.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Hour)
		synctest.Wait()
		if inspectIdle(s, p).session || inspectIdle(s, p).catalog || inspectIdle(s, p).timer {
			t.Fatal("shutdown retained idle state")
		}
	})
}

func TestConfigReloadBindingAndLogout(t *testing.T) {
	s, _ := persistentService(t)
	defer s.Shutdown(context.Background())
	ctx := context.Background()
	policy := config.Profile{Email: "test@example.invalid", LockTimeout: "15m"}
	if err := s.ApplyConfig(ctx, &config.File{Profiles: map[string]config.Profile{"default": policy}}); err != nil {
		t.Fatal(err)
	}
	saved, err := s.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if s.policies["default"].LockTimeout != "15m0s" {
		t.Fatal("policy not applied")
	}
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: policy.Email}
	if err := s.saveLogin(ctx, saved, cache, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	policy.Email = "different@example.invalid"
	if err := s.ApplyConfig(ctx, &config.File{Profiles: map[string]config.Profile{"default": policy}}); err == nil {
		t.Fatal("configured identity rebound cache")
	}
	recovered := &Service{Store: s.Store, Secrets: s.Secrets}
	saved, err = recovered.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered.policies) != 0 {
		t.Fatal("policy persisted in DB")
	}
	if s.policies["default"].Email != "test@example.invalid" {
		t.Fatal("failed reload changed policy")
	}
	if err := s.ApplyConfig(ctx, nil); err == nil {
		t.Fatal("removed logged-in profile")
	}
	if _, err := s.Logout(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	saved, err = s.load(ctx, "default")
	if err != nil || len(s.policies) == 0 || saved.Cache != nil {
		t.Fatal("logout should retain config only", err)
	}
	if err := s.ApplyConfig(ctx, nil); err != nil {
		t.Fatal(err)
	}
}

func TestIdlePolicyChangeDoesNotResetActivity(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		defer s.Shutdown(context.Background())
		p := idleProfile(t, s, "default", "10s")
		time.Sleep(6 * time.Second)
		p.gate <- struct{}{}
		s.mu.Lock()
		p.idleTimeout = 5 * time.Second
		s.scheduleIdle(p)
		s.mu.Unlock()
		<-p.gate
		synctest.Wait()
		if inspectIdle(s, p).session {
			t.Fatal("shortened policy reset activity instead of expiring")
		}
		p = idleProfile(t, s, "default", "1s")
		p.gate <- struct{}{}
		s.mu.Lock()
		p.idleTimeout = 0
		s.scheduleIdle(p)
		s.mu.Unlock()
		<-p.gate
		time.Sleep(time.Hour)
		synctest.Wait()
		if !inspectIdle(s, p).session || inspectIdle(s, p).timer {
			t.Fatal("disabled timeout still fired")
		}
	})
}

type idleState struct {
	session, catalog, timer bool
	barriers                int
	version                 uint64
}

func inspectIdle(s *Service, p *profile) idleState {
	p.gate <- struct{}{}
	defer func() { <-p.gate }()
	s.mu.Lock()
	defer s.mu.Unlock()
	return idleState{p.session != nil, p.catalog != nil, p.idleTimer != nil, p.barriers, p.idleVersion}
}

func TestReloadIsAtomicAndCancellationLeavesPolicyUnchanged(t *testing.T) {
	s, _ := persistentService(t)
	defer s.Shutdown(context.Background())
	ctx := context.Background()
	policy := config.Profile{Email: "test@example.invalid", LockTimeout: "15m"}
	initial := &config.File{Profiles: map[string]config.Profile{"a": policy, "b": policy}}
	if err := s.ApplyConfig(ctx, initial); err != nil {
		t.Fatal(err)
	}
	saved, err := s.load(ctx, "b")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.saveLogin(ctx, saved, &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: policy.Email}, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	shorter := policy
	shorter.LockTimeout = "1s"
	wrong := policy
	wrong.Email = "wrong@example.invalid"
	if err := s.ApplyConfig(ctx, &config.File{Profiles: map[string]config.Profile{"a": shorter, "b": wrong}}); err == nil {
		t.Fatal("accepted identity change")
	}
	if s.policies["a"].LockTimeout != "15m0s" {
		t.Fatal("partially applied config")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := s.ApplyConfig(canceled, &config.File{Profiles: map[string]config.Profile{"a": shorter, "b": policy}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if s.policies["a"].LockTimeout != "15m0s" {
		t.Fatal("canceled reload changed config")
	}
}

func TestReloadDoesNotPreventLockCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		defer s.Shutdown(context.Background())
		ctx, _, done, err := s.begin(context.Background(), "default", true)
		if err != nil {
			t.Fatal(err)
		}
		reloaded := make(chan error, 1)
		go func() { reloaded <- s.ApplyConfig(context.Background(), nil) }()
		locked := make(chan error, 1)
		go func() { _, err := s.Lock(context.Background(), ProfileRequest{Profile: "default"}); locked <- err }()
		<-ctx.Done()
		if ctx.Err() == nil {
			t.Fatal("reload prevented cancellation")
		}
		done()
		if err := <-reloaded; err != nil {
			t.Fatal(err)
		}
		if err := <-locked; err != nil {
			t.Fatal(err)
		}
	})
}

func TestStatusReportsAppliedPolicyWithoutRenewingDeadline(t *testing.T) {
	s, _ := persistentService(t)
	defer s.Shutdown(context.Background())
	policy := config.Profile{Email: "test@example.invalid", LockTimeout: "15m"}
	if err := s.ApplyConfig(t.Context(), &config.File{Profiles: map[string]config.Profile{"default": policy}}); err != nil {
		t.Fatal(err)
	}
	saved, err := s.load(t.Context(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.saveLogin(t.Context(), saved, &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: policy.Email}, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	p := idleProfile(t, s, "default", "15m")
	s.mu.Lock()
	deadline, activity := p.idleDeadline, p.lastActivity
	s.mu.Unlock()
	status, err := s.Status(t.Context(), ProfileRequest{Profile: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if status.State != "unlocked" || status.Server != "https://example.invalid" || status.Email != policy.Email || status.LockTimeout != "15m0s" || status.LockAt != deadline.UTC().Format(time.RFC3339Nano) {
		t.Fatal(status)
	}
	s.mu.Lock()
	unchanged := p.idleDeadline.Equal(deadline) && p.lastActivity.Equal(activity)
	s.mu.Unlock()
	if !unchanged {
		t.Fatal("status renewed activity")
	}
	if _, err := s.Lock(t.Context(), ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	status, err = s.Status(t.Context(), ProfileRequest{Profile: "default"})
	if err != nil || status.State != "locked" || status.LockAt != "" {
		t.Fatal(status, err)
	}
}

func TestBackgroundListDoesNotRenewIdleButActivityDoes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		defer s.Shutdown(context.Background())
		p := idleProfile(t, s, "default", "10s")
		p.catalogID = "generation"
		time.Sleep(4 * time.Second)
		if _, err := s.List(t.Context(), ListRequest{Profile: "default", Kind: "item", Background: true}); err != nil {
			t.Fatal(err)
		}
		s.mu.Lock()
		deadline := p.idleDeadline
		s.mu.Unlock()
		if remaining := time.Until(deadline); remaining != 6*time.Second {
			t.Fatal("background renewed deadline", remaining)
		}
		if _, err := s.Activity(t.Context(), ProfileRequest{Profile: "default"}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(6 * time.Second)
		synctest.Wait()
		if !inspectIdle(s, p).session {
			t.Fatal("activity did not renew")
		}
		time.Sleep(4 * time.Second)
		synctest.Wait()
		if inspectIdle(s, p).session {
			t.Fatal("background kept session alive")
		}
	})
}

func TestReloadCancellationBeforeReading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		s.configMu.Lock()
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		started := make(chan struct{})
		go func() { close(started); done <- s.ReloadConfig(ctx, path) }()
		<-started
		cancel()
		s.configMu.Unlock()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatalf("read despite cancellation: %v", err)
		}
	})
}

func TestReloadInvalidFileKeepsAppliedConfig(t *testing.T) {
	s, _ := persistentService(t)
	defer func() {
		if err := s.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("version = 1\n[profiles.default]\nemail = 'test@example.invalid'\nlock_timeout = '15m'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.ReloadConfig(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("invalid TOML"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.ReloadConfig(t.Context(), path); err == nil {
		t.Fatal("accepted invalid config")
	}
	if s.policies["default"].LockTimeout != "15m0s" {
		t.Fatal("lost applied policy")
	}
}
