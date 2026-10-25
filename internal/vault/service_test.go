package vault

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/134130/bwb/internal/vaultapi"
)

func TestLockCancelsActiveAndRejectsQueuedOperation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		ctx, _, finish, err := s.begin(context.Background(), "default", true)
		if err != nil {
			t.Fatal(err)
		}
		queued := make(chan error, 1)
		go func() {
			_, _, done, err := s.begin(context.Background(), "default", true)
			if done != nil {
				done()
			}
			queued <- err
		}()
		synctest.Wait()
		locked := make(chan error, 1)
		go func() {
			_, err := s.Lock(context.Background(), vaultapi.ProfileRequest{Profile: "default"})
			locked <- err
		}()
		synctest.Wait()
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatal("active operation was not canceled")
		}
		select {
		case <-locked:
			t.Fatal("lock returned before active operation released its keys")
		default:
		}
		finish()
		synctest.Wait()
		if err := <-queued; !errors.Is(err, errLocked) {
			t.Fatal("queued operation survived lock", err)
		}
		if err := <-locked; err != nil {
			t.Fatal(err)
		}
		status, err := s.Status(context.Background(), vaultapi.ProfileRequest{Profile: "default"})
		if err != nil || status.Unlocked {
			t.Fatal("profile unlocked after lock", err)
		}
	})
}

func TestShutdownCancelsAllProfilesAndPreventsReopen(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		ctxA, _, doneA, _ := s.begin(context.Background(), "a", true)
		ctxB, _, doneB, _ := s.begin(context.Background(), "b", true)
		stopped := make(chan error, 1)
		go func() { stopped <- s.Shutdown(context.Background()) }()
		synctest.Wait()
		if ctxA.Err() == nil || ctxB.Err() == nil {
			t.Fatal("shutdown did not cancel all operations")
		}
		doneA()
		doneB()
		synctest.Wait()
		if err := <-stopped; err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := s.begin(context.Background(), "a", true); !errors.Is(err, errLocked) {
			t.Fatal("shutdown allowed login", err)
		}
	})
}

func TestQueuedCancellationAndProfileValidation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		_, _, done, _ := s.begin(context.Background(), "default", true)
		defer done()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() { _, _, _, err := s.begin(ctx, "default", false); result <- err }()
		synctest.Wait()
		cancel()
		synctest.Wait()
		if err := <-result; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
	s := &Service{}
	for _, name := range []string{"", "..", "../other", "a/b"} {
		if _, err := s.Status(context.Background(), vaultapi.ProfileRequest{Profile: name}); err == nil {
			t.Fatal("invalid profile accepted")
		}
	}
	if _, err := s.Read(context.Background(), vaultapi.ReadRequest{Profile: "default"}); !errors.Is(err, errLocked) {
		t.Fatal(err)
	}
	if _, err := s.Sync(context.Background(), vaultapi.ProfileRequest{Profile: "default"}); !errors.Is(err, errLocked) {
		t.Fatal(err)
	}
	if len(s.profiles) != 0 {
		t.Fatal("read created state")
	}
}
