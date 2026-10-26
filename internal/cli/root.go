package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/daemon"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{Use: "bwb", Version: daemon.Options().Version, Short: "Browse and read Bitwarden vaults through a local daemon", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(&cobra.Command{Use: "version", Short: "Show this binary's version without starting the daemon", Args: cobra.NoArgs,
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "bwb version %s\n", root.Version)
			return err
		}})
	var profile string
	var configPath string
	root.PersistentFlags().StringVar(&configPath, "config", "", "Config file (overrides BWB_CONFIG)")
	var nonInteractive bool
	root.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Never prompt or automatically update the daemon")
	root.PersistentFlags().StringVar(&profile, "profile", "default", "Vault profile")
	var server, email string
	var passwordStdin bool
	var apiKeyLogin bool
	login := &cobra.Command{Use: "login", Short: "Log in and cache an encrypted snapshot", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if server != "" && (email != "" || apiKeyLogin) {
				if _, err := bitwarden.NewClient(server, email, uuid.NewString()); err != nil {
					return err
				}
			}
			var apiKey bitwarden.APIKey
			if apiKeyLogin {
				var err error
				apiKey, err = commandAPIKey(cmd, nonInteractive, authPrompt{Action: "login", Profile: profile, Server: server, Email: email})
				defer clear(apiKey.ClientSecret)
				if err != nil {
					return err
				}
			}
			password, err := commandPassword(cmd, passwordStdin, nonInteractive, authPrompt{Action: "login", Profile: profile, Server: server, Email: email})
			if err != nil {
				return err
			}
			defer clear(password)
			client, err := connectCommand(cmd)
			if err != nil {
				return err
			}
			defer client.Close()
			req := vaultapi.LoginRequest{Profile: profile, Server: server, Email: email, Password: password}
			if apiKeyLogin {
				_, err = client.Invoke(cmd.Context(), vaultapi.LoginAPIKey, vaultapi.LoginAPIKeyRequest{LoginRequest: req, APIKey: apiKey})
				var rpc *kit.RPCError
				if errors.As(err, &rpc) && rpc.Code == -32601 {
					return errors.New("API key login requires an updated daemon; run daemon update")
				}
			} else {
				_, err = client.Invoke(cmd.Context(), vaultapi.Login, req)
			}
			return err
		}}
	login.Flags().StringVar(&server, "server", "", "Vault URL (HTTPS, or loopback HTTP)")
	login.Flags().StringVar(&email, "email", "", "Account email (optional with --apikey)")
	login.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Read master password from stdin")
	login.Flags().BoolVar(&apiKeyLogin, "apikey", false, "Authenticate with a personal API key, then unlock with the master password")
	var unlockStdin bool
	unlock := &cobra.Command{Use: "unlock", Short: "Unlock a saved vault without network access", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		password, err := commandPassword(cmd, unlockStdin, nonInteractive, authPrompt{Action: "unlock", Profile: profile})
		if err != nil {
			return err
		}
		defer clear(password)
		client, err := connectCommand(cmd)
		if err != nil {
			return err
		}
		defer client.Close()
		_, err = client.Invoke(cmd.Context(), vaultapi.Unlock, vaultapi.UnlockRequest{Profile: profile, Password: password})
		return err
	}}
	unlock.Flags().BoolVar(&unlockStdin, "password-stdin", false, "Read master password from stdin")
	logout := &cobra.Command{Use: "logout", Short: "Delete saved credentials and encrypted cache for this profile", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := connectCommand(cmd)
		if err != nil {
			return err
		}
		defer client.Close()
		_, err = client.Invoke(cmd.Context(), vaultapi.Logout, vaultapi.ProfileRequest{Profile: profile})
		return err
	}}
	sync := &cobra.Command{Use: "sync", Short: "Refresh the encrypted snapshot", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connectCommand(cmd)
			if err != nil {
				return err
			}
			defer client.Close()
			execution := executor{Client: client, Prompt: interactivePrompt(cmd, nonInteractive)}
			_, err = execution.Sync(cmd.Context(), vaultapi.ProfileRequest{Profile: profile})
			return err
		}}
	lock := &cobra.Command{Use: "lock", Short: "Destroy in-memory keys while retaining the saved login", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := daemon.Manager()
			if err != nil {
				return err
			}
			status, err := manager.Status(cmd.Context())
			if err != nil {
				return err
			}
			if status.State == "stopped" {
				return nil
			}
			client, err := kit.Connect(cmd.Context(), daemon.Options())
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.Invoke(cmd.Context(), vaultapi.Lock, vaultapi.ProfileRequest{Profile: profile})
			return err
		}}
	var verbose bool
	status := &cobra.Command{Use: "status", Short: "Show profile state without starting the daemon", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := daemon.Manager()
			if err != nil {
				return err
			}
			state, err := manager.Status(cmd.Context())
			if err != nil {
				return err
			}
			if state.State == "stopped" {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "stopped")
				return err
			}
			client, err := kit.Connect(cmd.Context(), daemon.Options())
			if err != nil {
				return err
			}
			defer client.Close()
			result, err := client.Invoke(cmd.Context(), vaultapi.Status, vaultapi.ProfileRequest{Profile: profile})
			if err != nil {
				return err
			}
			if verbose {
				return writeStatus(cmd.OutOrStdout(), profile, result)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), result.State)
			return err
		}}
	status.Flags().BoolVar(&verbose, "verbose", false, "Show applied identity, lock timeout and deadline")
	control := &cobra.Command{Use: "daemon", Short: "Manage the local daemon"}
	control.AddCommand(daemonStatusCommand())
	control.AddCommand(&cobra.Command{Use: "run", Short: "Run the daemon in the foreground", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return daemon.Run(cmd.Context(), configPath) }})
	control.AddCommand(&cobra.Command{Use: "reload", Short: "Validate and apply the running daemon's config file", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := kit.Connect(cmd.Context(), daemon.Options())
		if err != nil {
			return err
		}
		defer client.Close()
		_, err = client.Invoke(cmd.Context(), vaultapi.Reload, vaultapi.Empty{})
		return err
	}})
	control.AddCommand(&cobra.Command{Use: "stop", Short: "Stop the daemon and destroy all sessions", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := daemon.Manager()
			if err != nil {
				return err
			}
			return manager.Stop(cmd.Context())
		}})
	control.AddCommand(&cobra.Command{Use: "update", Short: "Select this executable as the daemon release (restarts a running daemon)", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		manager, err := daemon.Manager()
		if err != nil {
			return err
		}
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		return manager.Update(cmd.Context(), executable)
	}})
	for _, kind := range []string{"item", "folder", "organization", "collection"} {
		root.AddCommand(listCommand(kind, &profile, &nonInteractive))
	}
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// Any command can launch or update the daemon; unused API keys must not be inherited.
		if cmd != login || !apiKeyLogin {
			if err := errors.Join(os.Unsetenv("BWB_CLIENTID"), os.Unsetenv("BWB_CLIENTSECRET")); err != nil {
				return err
			}
		}
		return prepareConfig(cmd, configPath, &profile, root.PersistentFlags().Changed("profile"))
	}
	configuration, profiles := configCommands()
	root.AddCommand(itemCommand(true, &profile, &nonInteractive))
	root.AddCommand(configuration, profiles, tuiCommand(&profile, &nonInteractive))
	root.AddCommand(login, unlock, logout, sync, lock, status, control)
	return root
}

func interactivePrompt(cmd *cobra.Command, disabled bool) passwordPrompt {
	if disabled {
		return nil
	}
	input, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(input.Fd())) {
		return nil
	}
	return terminalPassword
}

func commandPassword(cmd *cobra.Command, stdin, disabled bool, request authPrompt) ([]byte, error) {
	if stdin {
		return passwordInput(cmd)
	}
	prompt := interactivePrompt(cmd, disabled)
	if prompt == nil {
		return nil, errors.New("--password-stdin is required in non-interactive mode")
	}
	return prompt(cmd.Context(), request)
}

func passwordInput(cmd *cobra.Command) ([]byte, error) {
	password, err := readPassword(cmd.Context(), cmd.InOrStdin())
	if err != nil {
		return nil, err
	}
	input := password
	if bytes.HasSuffix(input, []byte("\n")) {
		input = bytes.TrimSuffix(input, []byte("\n"))
		input = bytes.TrimSuffix(input, []byte("\r"))
	}
	if len(input) == 0 || len(input) > 4096 {
		clear(password)
		return nil, errors.New("invalid password length")
	}
	// Erase the removed newline bytes as well as the password when its caller finishes.
	clear(password[len(input):])
	return input, nil
}

// shortcut: uninterruptible readers live until EOF or CLI exit; use cancellable readers for future interactive input.
func readPassword(ctx context.Context, input io.Reader) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if closer, ok := input.(io.ReadCloser); ok {
		// Inherited OS stdin can block both Read and Close; cancellation must not wait for either.
		stop := context.AfterFunc(ctx, func() { _ = closer.Close() })
		defer stop()
	}
	type result struct {
		password []byte
		err      error
	}
	ready := make(chan result)
	go func() {
		password, err := io.ReadAll(io.LimitReader(input, 4099))
		select {
		case ready <- result{password, err}:
		case <-ctx.Done():
			clear(password)
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case read := <-ready:
		if err := ctx.Err(); err != nil {
			clear(read.password)
			return nil, err
		}
		if read.err != nil {
			clear(read.password)
			return nil, errors.New("cannot read password")
		}
		return read.password, nil
	}
}
