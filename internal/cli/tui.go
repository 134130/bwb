package cli

import (
	"errors"
	"os"

	"github.com/134130/bwb/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func tuiCommand(profile *string, nonInteractive *bool) *cobra.Command {
	return &cobra.Command{Use: "tui", Short: "Explore vaults in an interactive terminal", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if *nonInteractive {
			return errors.New("tui requires interactive authentication")
		}
		input, inOK := cmd.InOrStdin().(*os.File)
		output, outOK := cmd.OutOrStdout().(*os.File)
		if !inOK || !outOK || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
			return errors.New("tui requires terminal input and output")
		}
		client, err := connectCommand(cmd)
		if err != nil {
			return err
		}
		defer client.Close()
		return tui.Run(cmd.Context(), *profile, client, input, output)
	}}
}
