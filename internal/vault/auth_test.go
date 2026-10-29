package vault

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
)

func authService(t *testing.T) *Service {
	t.Helper()
	s, _ := persistentService(t)
	r, err := s.load(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.saveLogin(context.Background(), r, &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "user@example.invalid"}, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAuthReservationWaitsForOneOwner(t *testing.T) {
	s := authService(t)
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		request := AuthRequest{Profile: "default", Action: "unlock"}
		first, err := s.BeginAuth(ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		result := make(chan AuthResponse, 1)
		go func() {
			r, err := s.BeginAuth(ctx, request)
			if err != nil {
				t.Error(err)
			}
			result <- r
		}()
		synctest.Wait()
		select {
		case <-result:
			t.Fatal("second client acquired prompt")
		default:
		}
		_, p, done, err := s.begin(ctx, "default", true)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.claimAuth(ctx, p, first.ID, "unlock"); err != nil {
			t.Fatal(err)
		}
		s.mu.Lock()
		p.authVersion++
		p.auth.succeeded = true
		s.finishAuth(p)
		s.mu.Unlock()
		done()
		synctest.Wait()
		if r := <-result; !r.Ready || r.ID != "" {
			t.Fatal("waiter requested another prompt")
		}
		// A very fast prompt can complete before the watch RPC reaches the daemon.
		watch, err := s.WatchAuth(ctx, AuthEndRequest{Profile: "default", ID: first.ID})
		if err != nil || !watch.Succeeded {
			t.Fatal("late watcher lost outcome", err)
		}
	})
}

func TestLockAndOwnerCancellationInvalidatePromptWithoutCascade(t *testing.T) {
	for _, lock := range []bool{false, true} {
		t.Run(map[bool]string{false: "owner canceled", true: "locked"}[lock], func(t *testing.T) {
			s := authService(t)
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				req := AuthRequest{Profile: "default", Action: "unlock"}
				first, err := s.BeginAuth(ctx, req)
				if err != nil {
					t.Fatal(err)
				}
				watchCtx, cancel := context.WithCancel(ctx)
				defer cancel()
				watching := make(chan error, 1)
				go func() {
					_, err := s.WatchAuth(watchCtx, AuthEndRequest{Profile: "default", ID: first.ID})
					watching <- err
				}()
				waiting := make(chan error, 1)
				go func() { _, err := s.BeginAuth(ctx, req); waiting <- err }()
				synctest.Wait()
				if lock {
					if _, err := s.Lock(ctx, ProfileRequest{Profile: "default"}); err != nil {
						t.Fatal(err)
					}
				} else {
					cancel()
				}
				synctest.Wait()
				if err := <-waiting; !errors.Is(err, ErrStateChanged) {
					t.Fatal("waiter started a second prompt", err)
				}
				<-watching
				if _, err := s.Unlock(ctx, UnlockRequest{Profile: "default", Password: []byte("late password"), AuthID: first.ID}); !errors.Is(err, ErrStateChanged) {
					t.Fatal("stale prompt could unlock", err)
				}
			})
		})
	}
}

func TestAbandonedReservationExpires(t *testing.T) {
	s := authService(t)
	synctest.Test(t, func(t *testing.T) {
		first, err := s.BeginAuth(context.Background(), AuthRequest{Profile: "default", Action: "unlock"})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(AuthTimeout)
		synctest.Wait()
		if _, err := s.Unlock(context.Background(), UnlockRequest{Profile: "default", AuthID: first.ID}); !errors.Is(err, ErrStateChanged) {
			t.Fatal(err)
		}
		if s.profiles["default"].auth != nil {
			t.Fatal("reservation leaked")
		}
	})
}

func TestFailedAuthenticationDoesNotCancelItsRPCReply(t *testing.T) {
	s := authService(t)
	synctest.Test(t, func(t *testing.T) {
		reservation, err := s.BeginAuth(t.Context(), AuthRequest{Profile: "default", Action: "unlock"})
		if err != nil {
			t.Fatal(err)
		}
		end := AuthEndRequest{Profile: "default", ID: reservation.ID}
		watching := make(chan AuthWatchResponse, 1)
		go func() {
			result, err := s.WatchAuth(t.Context(), end)
			if err != nil {
				t.Error(err)
			}
			watching <- result
		}()
		synctest.Wait()
		if _, err := s.Unlock(t.Context(), UnlockRequest{Profile: "default", AuthID: reservation.ID, Password: []byte("wrong")}); err == nil {
			t.Fatal("invalid cached vault unlocked")
		}
		synctest.Wait()
		if result := <-watching; !result.Failed || result.Succeeded {
			t.Fatal("RPC failure reported as external cancellation", result)
		}
		if result, err := s.WatchAuth(t.Context(), end); err != nil || !result.Failed || result.Succeeded {
			t.Fatal("late watcher lost failed RPC outcome", result, err)
		}
	})
}

func TestAuthRequiredDistinguishesMissingLoginAndLockedCache(t *testing.T) {
	s := authService(t)
	for _, name := range []string{"default", "absent"} {
		_, err := s.Read(context.Background(), ReadRequest{Profile: name})
		var auth *AuthRequired
		want := "unlock"
		if name == "absent" {
			want = "login"
		}
		if !errors.As(err, &auth) || auth.Action != want || auth.Profile != name {
			t.Fatal(err)
		}
	}
}
