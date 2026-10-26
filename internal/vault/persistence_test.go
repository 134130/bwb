package vault

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	daemon "github.com/134130/go-daemonkit"
	"github.com/134130/go-daemonkit/secrets"
)

type readyService struct{ ready chan struct{} }

func (s *readyService) Startup(context.Context) error { close(s.ready); return nil }

// Fault injection models native writes that changed the OS store before reporting failure.
type faultyCredentials struct {
	values            map[string][]byte
	setErr, deleteErr error
	afterSet          func()
	beforeGet         func()
	getErr            error
}

func (f *faultyCredentials) Get(_ context.Context, key string) ([]byte, error) {
	if f.beforeGet != nil {
		f.beforeGet()
	}
	if f.getErr != nil {
		return nil, f.getErr
	}
	v, ok := f.values[key]
	if !ok {
		return nil, secrets.ErrNotFound
	}
	return append([]byte(nil), v...), nil
}
func (f *faultyCredentials) Set(_ context.Context, key string, value []byte) error {
	f.values[key] = append([]byte(nil), value...)
	if f.afterSet != nil {
		f.afterSet()
	}
	return f.setErr
}
func (f *faultyCredentials) Delete(_ context.Context, key string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.values, key)
	return nil
}

func persistentService(t *testing.T) (*Service, *faultyCredentials) {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "bwb-store-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	t.Setenv("XDG_STATE_HOME", dir)
	app, err := daemon.New(daemon.Options{AppID: "bwb-test", Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	store, err := app.Storage("vault")
	if err != nil {
		t.Fatal(err)
	}
	ready := &readyService{ready: make(chan struct{})}
	if err := app.Register(ready); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- app.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("storage app did not stop")
		}
	})
	select {
	case <-ready.ready:
	case err := <-result:
		result <- err
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("storage did not open")
	}
	f := &faultyCredentials{values: map[string][]byte{}}
	return &Service{Store: store, Secrets: f}, f
}

func TestCredentialJournalAndLogoutFailure(t *testing.T) {
	s, f := persistentService(t)
	ctx := context.Background()
	r, _ := s.load(ctx, "default")
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "test@example.invalid"}
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("old")}); err != nil {
		t.Fatal(err)
	}
	old := r.Credential
	f.setErr = secrets.ErrAccess
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("replacement")}); !errors.Is(err, secrets.ErrAccess) {
		t.Fatal(err)
	}
	// Simulate process loss: only the durable record and native store survive.
	recovered := &Service{Store: s.Store, Secrets: f}
	saved, err := recovered.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Credential != old || len(saved.Garbage) != 1 || len(f.values) != 2 {
		t.Fatal("uncertain write lost recovery information")
	}
	token, err := recovered.loadCredential(ctx, saved)
	if err != nil || string(token.Token) != "old" {
		t.Fatal("failed login replaced valid credential", err)
	}
	token.clear()
	f.setErr = nil
	if err := recovered.saveLogin(ctx, saved, cache, credential{Token: []byte("new")}); err != nil {
		t.Fatal(err)
	}
	if len(f.values) != 2 || len(saved.Garbage) != 1 || saved.Garbage[0] != old {
		t.Fatal("journal did not collect abandoned credential")
	}
	f.deleteErr = secrets.ErrPermission
	if _, err := recovered.Logout(ctx, ProfileRequest{Profile: "default"}); !errors.Is(err, secrets.ErrPermission) {
		t.Fatal(err)
	}
	saved, err = recovered.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Deleting || saved.Cache != nil || len(saved.Garbage) != 2 {
		t.Fatal("failed logout did not revoke cached unlock")
	}
	if _, err := recovered.Unlock(ctx, UnlockRequest{Profile: "default", Password: []byte("irrelevant")}); err == nil {
		t.Fatal("incomplete logout allowed unlock")
	}
	f.deleteErr = nil
	if _, err := recovered.Logout(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if len(f.values) != 0 {
		t.Fatal("logout left credentials")
	}
	keys, err := s.Store.List(ctx, "")
	if err != nil || len(keys) != 0 {
		t.Fatal("logout left cached records", err)
	}
}

func TestCacheChunksAndIdentityBinding(t *testing.T) {
	s, f := persistentService(t)
	ctx := context.Background()
	r, _ := s.load(ctx, "a")
	cache := &bitwarden.Cache{Version: 1, DeviceID: strings.Repeat("x", 2*chunkSize), Server: "https://example.invalid", Email: "a@example.invalid"}
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := s.load(ctx, "a")
	if err != nil || reloaded.Cache.DeviceID != cache.DeviceID {
		t.Fatal("chunked cache mismatch", err)
	}
	// Copying a profile's record cannot silently select another profile's secret.
	if err := s.put(ctx, "b", reloaded); err != nil {
		t.Fatal(err)
	}
	if _, err := s.load(ctx, "b"); err == nil {
		t.Fatal("cross-profile record accepted")
	}
	reloaded.Cache.Email = "b@example.invalid"
	if _, err := s.loadCredential(ctx, reloaded); err == nil {
		t.Fatal("credential identity mismatch accepted")
	}
	if len(f.values) != 1 {
		t.Fatal("unexpected secret mutation")
	}
	if err := s.put(ctx, "a", nil); err != nil {
		t.Fatal(err)
	}
	keys, err := s.Store.List(ctx, "a/")
	if err != nil || len(keys) != 0 {
		t.Fatal("old chunks not deleted", err)
	}
}

func TestCanceledCommitLeavesRecoverableCredentialIntent(t *testing.T) {
	s, f := persistentService(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := s.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	f.afterSet = cancel
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("temporary")}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	saved, err := s.load(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Cache != nil || saved.Credential != "" || len(saved.Garbage) != 1 || len(f.values) != 1 {
		t.Fatal("canceled commit lost credential cleanup intent")
	}
	f.afterSet = nil
	if _, err := s.Logout(context.Background(), ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	if len(f.values) != 0 {
		t.Fatal("abandoned credential survived logout")
	}
}

func TestStaleMemorySessionCannotOverwriteCommittedCredential(t *testing.T) {
	s, f := persistentService(t)
	ctx := context.Background()
	r, err := s.load(ctx, "default")
	if err != nil {
		t.Fatal(err)
	}
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("old")}); err != nil {
		t.Fatal(err)
	}
	_, p, done, err := s.begin(ctx, "default", true)
	if err != nil {
		t.Fatal(err)
	}
	p.session = &bitwarden.Session{}
	p.credential = r.Credential
	done()
	// A commit can succeed even when cancellation prevents publishing the new memory session.
	if err := s.saveLogin(ctx, r, cache, credential{Token: []byte("new")}); err != nil {
		t.Fatal(err)
	}
	var auth *AuthRequired
	if _, err := s.Sync(ctx, ProfileRequest{Profile: "default"}); !errors.As(err, &auth) || auth.Action != "unlock" {
		t.Fatal(err)
	}
	token, err := s.loadCredential(ctx, r)
	if err != nil || string(token.Token) != "new" {
		t.Fatal("stale session overwrote credential", err)
	}
	token.clear()
	if len(f.values) != 2 {
		t.Fatal("stale session changed cleanup journal")
	}
}
