package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"text/tabwriter"

	"github.com/134130/bwb/internal/config"
	"github.com/134130/bwb/internal/daemon"
	kit "github.com/134130/go-daemonkit"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type commandConfig struct {
	path        string
	startupPath string
	file        *config.File
	interactive bool
	notice      func(string)
}
type configKey struct{}

func prepareConfig(cmd *cobra.Command, pathFlag string, profile *string, changed bool) error {
	// Daemon control remains available even with an invalid user configuration.
	if cmd.Parent() != nil && cmd.Parent().Name() == "daemon" {
		return nil
	}
	path, explicit, err := config.Path(pathFlag)
	if err != nil {
		return err
	}
	state := commandConfig{path: path}
	input, inOK := cmd.InOrStdin().(*os.File)
	output, outOK := cmd.OutOrStdout().(*os.File)
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	state.interactive = !nonInteractive && inOK && outOK && term.IsTerminal(int(input.Fd())) && term.IsTerminal(int(output.Fd()))
	var once sync.Once
	state.notice = func(message string) {
		once.Do(func() { _, _ = fmt.Fprintln(cmd.ErrOrStderr(), message) })
	}
	if explicit {
		state.startupPath = path
	}
	isConfig := cmd.Parent() != nil && cmd.Parent().Name() == "config"
	isProfile := cmd.Parent() != nil && cmd.Parent().Name() == "profile"
	if !(isConfig && cmd.Name() == "path") && !((changed || os.Getenv("BWB_PROFILE") != "") && !isConfig && !isProfile) {
		state.file, err = config.Load(path, explicit)
		if err != nil {
			return err
		}
	}
	if !isConfig && !isProfile {
		*profile, err = config.Select(*profile, changed, state.file)
		if err != nil {
			return err
		}
	}
	cmd.SetContext(context.WithValue(cmd.Context(), configKey{}, state))
	return nil
}

func connectCommand(cmd *cobra.Command) (*kit.Client, error) { return connectContext(cmd.Context()) }

func connectContext(ctx context.Context) (*kit.Client, error) {
	state, _ := ctx.Value(configKey{}).(commandConfig)
	return daemon.Connect(ctx, state.startupPath, state.interactive, state.notice)
}

func configCommands() (*cobra.Command, *cobra.Command) {
	configuration := &cobra.Command{Use: "config", Short: "Inspect and validate non-secret configuration"}
	configuration.AddCommand(&cobra.Command{Use: "path", Args: cobra.NoArgs, Short: "Show the selected config path", RunE: func(cmd *cobra.Command, _ []string) error {
		state, _ := cmd.Context().Value(configKey{}).(commandConfig)
		_, err := fmt.Fprintln(cmd.OutOrStdout(), state.path)
		return err
	}})
	configuration.AddCommand(&cobra.Command{Use: "validate", Args: cobra.NoArgs, Short: "Validate TOML, profiles and lock durations", RunE: func(cmd *cobra.Command, _ []string) error {
		state, _ := cmd.Context().Value(configKey{}).(commandConfig)
		if state.file == nil {
			return errors.New("config file does not exist")
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "config valid")
		return err
	}})
	profiles := &cobra.Command{Use: "profile", Short: "Inspect configured profiles without starting the daemon"}
	write := func(cmd *cobra.Command, names []string) error {
		state, _ := cmd.Context().Value(configKey{}).(commandConfig)
		table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		if _, err := fmt.Fprintln(table, "NAME\tSERVER\tEMAIL\tLOCK_TIMEOUT"); err != nil {
			return err
		}
		for _, name := range names {
			p := state.file.Profiles[name]
			if _, err := fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", display(name), display(p.Server), display(p.Email), p.LockTimeout); err != nil {
				return err
			}
		}
		return table.Flush()
	}
	profiles.AddCommand(&cobra.Command{Use: "list", Args: cobra.NoArgs, Short: "List profiles defined in config", RunE: func(cmd *cobra.Command, _ []string) error {
		state, _ := cmd.Context().Value(configKey{}).(commandConfig)
		names := []string{}
		if state.file != nil {
			for name := range state.file.Profiles {
				names = append(names, name)
			}
		}
		slices.Sort(names)
		return write(cmd, names)
	}})
	profiles.AddCommand(&cobra.Command{Use: "show NAME", Args: cobra.ExactArgs(1), Short: "Show one configured profile", RunE: func(cmd *cobra.Command, args []string) error {
		state, _ := cmd.Context().Value(configKey{}).(commandConfig)
		if state.file == nil {
			return errors.New("profile is not defined in config")
		}
		if _, ok := state.file.Profiles[args[0]]; !ok {
			return errors.New("profile is not defined in config")
		}
		return write(cmd, args)
	}})
	return configuration, profiles
}
