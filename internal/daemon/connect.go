package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"golang.org/x/mod/semver"
	"golang.org/x/sys/unix"
)

func newer(client, server string) bool {
	return semver.IsValid(client) && semver.IsValid(server) &&
		semver.Prerelease(client) == "" && semver.Prerelease(server) == "" &&
		semver.Compare(client, server) > 0
}

func Connect(ctx context.Context, interactive bool, notice func(string)) (*kit.Client, error) {
	manager, err := Manager()
	if err != nil {
		return nil, err
	}
	if err := manager.EnsureRunning(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect to daemon; run bwb daemon update if versions are incompatible: %w", err)
	}
	state, err := manager.Status(ctx)
	if err != nil {
		return nil, err
	}
	if newer(Options().Version, state.Version) {
		if !interactive {
			if notice != nil {
				notice(fmt.Sprintf("bwb: daemon %s is older than CLI %s. Run: bwb daemon update", state.Version, Options().Version))
			}
		} else if err := upgrade(ctx, manager, state.StateDir); err != nil {
			return nil, err
		}
	}
	return kit.Connect(ctx, Options())
}

func upgrade(ctx context.Context, manager *kit.Manager, directory string) error {
	unlock, err := lockUpgrade(ctx, directory)
	if err != nil {
		return err
	}
	defer unlock()
	// Another CLI may have upgraded while this process waited.
	state, err := manager.Status(ctx)
	if err != nil {
		return err
	}
	if state.State != kit.StatusReady || !newer(Options().Version, state.Version) {
		return nil
	}
	client, err := kit.Connect(ctx, Options())
	if err != nil {
		return err
	}
	status, err := client.Invoke(ctx, vaultapi.DaemonStatus, vaultapi.Empty{})
	client.Close()
	if err != nil {
		return fmt.Errorf("cannot preserve daemon configuration; run bwb daemon update: %w", err)
	}
	if !filepath.IsAbs(status.ConfigPath) {
		return errors.New("daemon returned an invalid config path")
	}
	// Validate the active path before stopping any sessions.
	if _, err := config.Load(status.ConfigPath, false); err != nil {
		return err
	}
	previous, existed := os.LookupEnv("BWB_CONFIG")
	if err := os.Unsetenv("BWB_CONFIG"); err != nil {
		return err
	}
	defer func() {
		if existed {
			_ = os.Setenv("BWB_CONFIG", previous)
		}
	}()
	args := []string{"daemon", "run", "--config", status.ConfigPath}
	if _, err := os.Stat(status.ConfigPath); errors.Is(err, os.ErrNotExist) {
		defaultPath, _, err := config.Path("")
		if err != nil || status.ConfigPath != defaultPath {
			return errors.New("active daemon config is missing; restore it before updating")
		}
		args = []string{"daemon", "run"}
	} else if err != nil {
		return err
	} else if err := os.Setenv("BWB_CONFIG", status.ConfigPath); err != nil {
		return err
	}
	// Preserve the path for rollback of older selections that used environment-only config.
	// Explicit argv preserves the daemon's path even when this CLI selected another file.
	manager, err = kit.NewManager(Options(), args...)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := manager.Update(ctx, executable); err != nil {
		return fmt.Errorf("daemon update failed: %w", err)
	}
	return nil
}

// The separate lock serializes bwb's automatic version decisions; daemonkit owns replacement.
func lockUpgrade(ctx context.Context, directory string) (func(), error) {
	fd, err := unix.Open(filepath.Join(directory, "auto-update.lock"), unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), "auto-update.lock")
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		file.Close()
		return nil, errors.New("invalid automatic update lock")
	}
	for {
		if err := ctx.Err(); err != nil {
			file.Close()
			return nil, err
		}
		err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return func() { file.Close() }, nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EAGAIN) {
			file.Close()
			return nil, err
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
