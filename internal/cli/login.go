package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Resolve saved identity before asking for the fields needed by this login method.
func loginIdentity(cmd *cobra.Command, client *kit.Client, profile, server, email string, needEmail, nonInteractive bool) (string, string, error) {
	status, err := client.Invoke(cmd.Context(), vaultapi.Status, vaultapi.ProfileRequest{Profile: profile})
	if err != nil {
		return "", "", err
	}
	if server == "" {
		server = status.Server
	}
	if email == "" {
		email = status.Email
	}
	prompt := interactivePrompt(cmd, nonInteractive)
	for _, field := range []struct {
		value       *string
		label, flag string
		required    bool
	}{
		{&server, "Server URL", "--server", true},
		{&email, "Email", "--email", needEmail},
	} {
		for field.required && *field.value == "" {
			if prompt == nil {
				return "", "", fmt.Errorf("%s is required for first login in non-interactive mode", field.flag)
			}
			value, err := prompt(cmd.Context(), authPrompt{Action: "login", Profile: profile, Label: field.label, Public: true})
			if err != nil {
				clear(value)
				return "", "", err
			}
			*field.value = strings.TrimSpace(string(value))
			clear(value)
			validationEmail := email
			if field.flag == "--server" {
				validationEmail = ""
			}
			_, err = bitwarden.NewClient(server, validationEmail, uuid.NewString())
			if *field.value == "" {
				err = errors.New(field.label + " is required")
			}
			if err != nil {
				*field.value = ""
				if _, err := fmt.Fprintln(cmd.ErrOrStderr(), err); err != nil {
					return "", "", err
				}
			}
		}
	}
	_, err = bitwarden.NewClient(server, email, uuid.NewString())
	return server, email, err
}
