package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
func (f *faultyCredentials) Set(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(value) > 2560 {
		return secrets.ErrLimit
	}
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

func TestCredentialChunksSurviveRotationAndRestart(t *testing.T) {
	s, f := persistentService(t)
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}
	r, _ := s.load(t.Context(), "default")
	if err := s.saveLogin(t.Context(), r, cache, credential{Token: []byte("legacy")}); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{strings.Repeat("long-refresh.", 512), strings.Repeat("rotated-refresh.", 1024), "short-again"} {
		if err := s.saveToken(t.Context(), r, []byte(token)); err != nil {
			t.Fatal(err)
		}
		restored := &Service{Store: s.Store, Secrets: f}
		var err error
		r, err = restored.load(t.Context(), "default")
		if err != nil {
			t.Fatal(err)
		}
		auth, err := restored.loadCredential(t.Context(), r)
		if err != nil || string(auth.Token) != token {
			t.Fatal("restored credential differs", err)
		}
		auth.clear()
		encoded, _ := json.Marshal(r)
		if bytes.Contains(encoded, []byte(token)) || bytes.Contains(encoded, []byte(base64.StdEncoding.EncodeToString([]byte(token)))) {
			t.Fatal("token was written to cache metadata")
		}
		for _, value := range f.values {
			if len(value) > credentialChunkSize {
				t.Fatal("native credential limit exceeded")
			}
		}
	}
	if _, err := s.Logout(t.Context(), ProfileRequest{Profile: "default"}); err != nil || len(f.values) != 0 {
		t.Fatal("logout retained credential chunks", err)
	}
}

func TestPartialCredentialWriteKeepsPreviousTokenAndCleanupJournal(t *testing.T) {
	for _, cancelWrite := range []bool{false, true} {
		t.Run(map[bool]string{false: "uncertain native write", true: "canceled write"}[cancelWrite], func(t *testing.T) {
			s, f := persistentService(t)
			cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}
			r, _ := s.load(t.Context(), "default")
			original := strings.Repeat("original-token.", 512)
			if err := s.saveLogin(t.Context(), r, cache, credential{Token: []byte(original)}); err != nil {
				t.Fatal(err)
			}
			previous := r.Credential
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			writes := 0
			f.afterSet = func() {
				writes++
				if writes == 2 {
					if cancelWrite {
						cancel()
					} else {
						f.setErr = secrets.ErrAccess
					}
				}
			}
			err := s.saveToken(ctx, r, []byte(strings.Repeat("replacement.", 1024)))
			want := secrets.ErrAccess
			if cancelWrite {
				want = context.Canceled
			}
			if !errors.Is(err, want) {
				t.Fatal("partial write unexpectedly succeeded", err)
			}
			f.afterSet, f.setErr = nil, nil
			restored := &Service{Store: s.Store, Secrets: f}
			r, err = restored.load(t.Context(), "default")
			if err != nil || r.Credential != previous || len(r.Garbage) != 1 {
				t.Fatal("partial write lost previous credential or cleanup journal", err)
			}
			auth, err := restored.loadCredential(t.Context(), r)
			if err != nil || string(auth.Token) != original {
				t.Fatal("partial rotation mixed token generations", err)
			}
			auth.clear()
			if err := restored.saveToken(t.Context(), r, []byte(strings.Repeat("recovered.", 1024))); err != nil {
				t.Fatal("could not recover failed chunked write", err)
			}
			if _, err := restored.Logout(t.Context(), ProfileRequest{Profile: "default"}); err != nil || len(f.values) != 0 {
				t.Fatal("logout retained abandoned credential chunks", err)
			}
		})
	}
}

func TestCredentialChunkValidation(t *testing.T) {
	s, f := persistentService(t)
	cache := &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}
	r, _ := s.load(t.Context(), "default")
	if err := s.saveLogin(t.Context(), r, cache, credential{Token: bytes.Repeat([]byte("x"), maxCredentialSize)}); err == nil || len(f.values) != 0 || len(r.Garbage) != 0 {
		t.Fatal("oversized credential wrote native values or journal", err)
	}
	if err := s.saveLogin(t.Context(), r, cache, credential{Token: bytes.Repeat([]byte("x"), 4096)}); err != nil {
		t.Fatal(err)
	}
	keys, err := credentialKeys(r.Credential)
	if err != nil || len(keys) < 2 || r.Version != 4 {
		t.Fatal("large credential not chunked", err)
	}
	part := f.values[keys[0]]
	for _, value := range [][]byte{nil, part[:len(part)-1], append(bytes.Clone(part), 'x')} {
		f.values[keys[0]] = value
		if auth, err := s.loadCredential(t.Context(), r); err == nil {
			auth.clear()
			t.Fatal("accepted incomplete credential chunk")
		}
	}
	delete(f.values, keys[0])
	if _, err := s.loadCredential(t.Context(), r); !errors.Is(err, secrets.ErrNotFound) {
		t.Fatal("missing chunk did not fail", err)
	}
	f.values[keys[0]] = part
	base := r.Credential[:strings.LastIndex(r.Credential, "/")]
	for _, reference := range []string{"", "default", "default/invalid/2", r.Credential + "/0", base + "/1", base + "/02", base + "/-2", base + "/999999"} {
		if _, err := credentialKeys(reference); err == nil {
			t.Fatal("accepted malformed credential reference", reference)
		}
	}
}
