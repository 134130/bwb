package daemon

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestUpgradeVersionPolicy(t *testing.T) {
	for _, tc := range []struct {
		cli, daemon string
		want        bool
	}{
		{"v0.4.0", "v0.3.0", true}, {"v0.10.0", "v0.9.0", true},
		{"v0.3.0", "v0.4.0", false}, {"v0.3.0", "v0.3.0", false},
		{"dev", "v0.3.0", false}, {"v0.4.0", "dev", false},
		{"v0.4.0-rc.1", "v0.3.0", false}, {"v0.4.0", "v0.4.0-rc.1", false},
		{"v0.4.0+build", "v0.4.0", false}, {"bad\nversion", "v0.3.0", false},
	} {
		if got := newer(tc.cli, tc.daemon); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}
func TestUpgradeLockSerializesAndCancels(t *testing.T) {
	dir := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		release, err := lockUpgrade(t.Context(), dir)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		done := make(chan error, 1)
		go func() {
			unlock, err := lockUpgrade(ctx, dir)
			if err == nil {
				unlock()
			}
			done <- err
		}()
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("lock did not serialize")
		default:
		}
		time.Sleep(time.Second)
		if err := <-done; err != context.DeadlineExceeded {
			t.Fatal(err)
		}
		release()
		unlock, err := lockUpgrade(t.Context(), dir)
		if err != nil {
			t.Fatal(err)
		}
		unlock()
	})
}
