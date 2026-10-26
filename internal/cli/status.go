package cli

import (
	"fmt"
	"io"

	"github.com/134130/bwb/internal/daemon"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/spf13/cobra"
)

func writeStatus(out io.Writer, profile string, result vaultapi.StatusResponse) error {
	value := func(s string) string {
		if s == "" {
			return "-"
		}
		return display(s)
	}
	_, err := fmt.Fprintf(out, "PROFILE\t%s\nSTATE\t%s\nSERVER\t%s\nEMAIL\t%s\nLOCK_TIMEOUT\t%s\nLOCK_AT\t%s\n", display(profile), display(result.State), value(result.Server), value(result.Email), value(result.LockTimeout), value(result.LockAt))
	return err
}

func daemonStatusCommand() *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Show daemon state, version and selected config path without starting it", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		manager, err := daemon.Manager()
		if err != nil {
			return err
		}
		state, err := manager.Status(cmd.Context())
		if err != nil {
			return err
		}
		if state.State != kit.StatusReady {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), state.State)
			return err
		}
		client, err := kit.Connect(cmd.Context(), daemon.Options())
		if err != nil {
			return err
		}
		defer client.Close()
		result, err := client.Invoke(cmd.Context(), vaultapi.DaemonStatus, vaultapi.Empty{})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "STATE\t%s\nVERSION\t%s\nCONFIG\t%s\n", display(string(state.State)), display(state.Version), display(result.ConfigPath))
		return err
	}}
}
