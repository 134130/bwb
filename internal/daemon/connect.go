package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"golang.org/x/mod/semver"
)

func newer(client, server string) bool {
	return semver.IsValid(client) && semver.IsValid(server) &&
		semver.Prerelease(client) == "" && semver.Prerelease(server) == "" &&
		semver.Compare(client, server) > 0
}

func Connect(ctx context.Context, configPath string, interactive bool, notice func(string)) (*kit.Client, error) {
	manager, err := launchManager(configPath)
	if err != nil {
		return nil, err
	}
	for range 3 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// EnsureRunning also recovers pending lifecycle journals after a conflict.
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
			} else {
				err := upgrade(ctx, state.UpdateToken)
				if errors.Is(err, kit.ErrUpdateConflict) {
					continue
				}
				if err != nil {
					return nil, err
				}
			}
		}
		return kit.Connect(ctx, Options())
	}
	return nil, fmt.Errorf("daemon changed during update; retry or run bwb daemon update: %w", kit.ErrUpdateConflict)
}

// Keep existing environment inheritance without mutating the CLI process.
func launchEnvironment(configPath string) []string {
	env := []string{}
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "BWB_CONFIG=") {
			env = append(env, entry)
		}
	}
	if configPath != "" {
		env = append(env, "BWB_CONFIG="+configPath)
	}
	return env
}

func launchManager(configPath string) (*kit.Manager, error) {
	args := []string{"daemon", "run"}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	return kit.NewManagerWithLaunch(Options(), kit.LaunchOptions{Env: launchEnvironment(configPath)}, args...)
}

func upgrade(ctx context.Context, token kit.UpdateToken) error {
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

	path, err := upgradeConfigPath(status)
	if err != nil {
		return err
	}

	// The environment also preserves older env-only selections during this rollback.
	manager, err := launchManager(path)
	if err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := manager.UpdateIf(ctx, executable, token); err != nil {
		return fmt.Errorf("daemon update failed: %w", err)
	}
	return nil
}

func upgradeConfigPath(status vaultapi.DaemonStatusResponse) (string, error) {
	if status.ConfigOptional {
		defaultPath, err := config.DefaultPath()
		if err != nil {
			return "", err
		}
		if status.ConfigPath != defaultPath {
			return "", errors.New("active daemon default config differs from launcher; update manually")
		}
		return "", nil
	}
	if _, err := os.Stat(status.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("active daemon config is missing; restore it before updating")
	} else if err != nil {
		return "", err
	}
	return status.ConfigPath, nil
}

// Update validates the launch config before stopping the selected daemon.
func Update(ctx context.Context, configPath string) error {
	path, explicit, err := config.Path(configPath)
	if err != nil {
		return err
	}
	if _, err := config.Load(path, explicit); err != nil {
		return err
	}
	if explicit {
		configPath = path
	}
	manager, err := launchManager(configPath)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return manager.Update(ctx, executable)
}
