package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/daemon"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{Use: "bwb", Short: "Read Bitwarden vault fields through a local daemon", SilenceUsage: true, SilenceErrors: true}
	var profile string
	root.PersistentFlags().StringVar(&profile, "profile", "default", "In-memory vault profile")
	var server, email string
	var passwordStdin bool
	login := &cobra.Command{Use: "login", Short: "Log in and cache an encrypted snapshot", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !passwordStdin {
				return errors.New("--password-stdin is required; interactive login is not implemented yet")
			}
			if _, err := bitwarden.NewClient(server, email, uuid.NewString()); err != nil {
				return err
			}
			password, err := readPassword(cmd.Context(), cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer clear(password)
			input := password
			if bytes.HasSuffix(input, []byte("\n")) {
				input = bytes.TrimSuffix(input, []byte("\n"))
				input = bytes.TrimSuffix(input, []byte("\r"))
			}
			if len(input) == 0 || len(input) > 4096 {
				return errors.New("invalid password length")
			}
			client, err := daemon.Connect(cmd.Context())
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.Invoke(cmd.Context(), vaultapi.Login, vaultapi.LoginRequest{Profile: profile, Server: server, Email: email, Password: input})
			return err
		}}
	login.Flags().StringVar(&server, "server", "", "Vault URL (HTTPS, or loopback HTTP)")
	login.Flags().StringVar(&email, "email", "", "Account email")
	login.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Read master password from stdin")
	_ = login.MarkFlagRequired("server")
	_ = login.MarkFlagRequired("email")
	read := &cobra.Command{Use: "read ITEM_ID FIELD", Short: "Read a cached field without contacting the server", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if id, err := uuid.Parse(args[0]); err != nil || id == uuid.Nil || id.String() != args[0] {
				return errors.New("invalid item ID")
			}
			client, err := daemon.Connect(cmd.Context())
			if err != nil {
				return err
			}
			defer client.Close()
			result, err := client.Invoke(cmd.Context(), vaultapi.Read, vaultapi.ReadRequest{Profile: profile, ID: args[0], Field: args[1]})
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), result.Value)
			return err
		}}
	sync := &cobra.Command{Use: "sync", Short: "Refresh the encrypted snapshot", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := daemon.Connect(cmd.Context())
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.Invoke(cmd.Context(), vaultapi.Sync, vaultapi.ProfileRequest{Profile: profile})
			return err
		}}
	lock := &cobra.Command{Use: "lock", Short: "Destroy this profile's session and snapshot", Args: cobra.NoArgs,
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
			value := "locked"
			if result.Unlocked {
				value = "unlocked"
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
			return err
		}}
	control := &cobra.Command{Use: "daemon", Short: "Manage the local daemon"}
	control.AddCommand(&cobra.Command{Use: "run", Short: "Run the daemon in the foreground", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return daemon.Run(cmd.Context()) }})
	control.AddCommand(&cobra.Command{Use: "stop", Short: "Stop the daemon and destroy all sessions", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manager, err := daemon.Manager()
			if err != nil {
				return err
			}
			return manager.Stop(cmd.Context())
		}})
	root.AddCommand(login, read, sync, lock, status, control)
	return root
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
