package cli

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{Use: "bwb", Short: "Read Bitwarden vault fields through the API", SilenceUsage: true, SilenceErrors: true}
	var server, email string
	var passwordStdin bool
	read := &cobra.Command{
		Use: "read ITEM_ID FIELD", Short: "Log in, sync, and print one plaintext field",
		Long: "Log in and read a field (for example login/password). This initial client has no daemon or saved session. The password is read from stdin, never an argument.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !passwordStdin {
				return errors.New("--password-stdin is required; interactive login is not implemented yet")
			}
			if id, err := uuid.Parse(args[0]); err != nil || id == uuid.Nil || id.String() != args[0] {
				return errors.New("invalid item ID")
			}
			client, err := bitwarden.NewClient(server, email, uuid.NewString())
			if err != nil {
				return err
			}
			password, err := readPassword(cmd.Context(), cmd.InOrStdin())
			if err != nil {
				return err
			}
			defer clear(password)
			// Remove one terminal line ending, preserving spaces and other password bytes.
			if len(password) > 4098 {
				return errors.New("password input too long")
			}
			input := password
			if bytes.HasSuffix(input, []byte("\n")) {
				input = bytes.TrimSuffix(input, []byte("\n"))
				input = bytes.TrimSuffix(input, []byte("\r"))
			}
			if len(input) == 0 || len(input) > 4096 {
				return errors.New("invalid password length")
			}
			session, err := client.Login(cmd.Context(), input)
			if err != nil {
				return err
			}
			defer session.Close()
			vault, err := session.Sync(cmd.Context())
			if err != nil {
				return err
			}
			value, err := vault.Read(args[0], args[1])
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), value)
			return err
		},
	}
	read.Flags().StringVar(&server, "server", "", "Vault URL (HTTPS, or loopback HTTP)")
	read.Flags().StringVar(&email, "email", "", "Account email")
	read.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Read master password from stdin")
	// These names are declared directly above; Cobra reports missing values before RunE.
	_ = read.MarkFlagRequired("server")
	_ = read.MarkFlagRequired("email")
	root.AddCommand(read)
	return root
}

// shortcut: uninterruptible reads live until EOF or CLI exit; require cancellable input before daemon reuse.
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
