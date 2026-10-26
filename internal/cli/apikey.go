package cli

import (
	"errors"
	"os"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/spf13/cobra"
)

func commandAPIKey(cmd *cobra.Command, disabled bool, request authPrompt) (bitwarden.APIKey, error) {
	key := bitwarden.APIKey{ClientID: os.Getenv("BWB_CLIENTID"), ClientSecret: []byte(os.Getenv("BWB_CLIENTSECRET"))}
	if err := errors.Join(os.Unsetenv("BWB_CLIENTID"), os.Unsetenv("BWB_CLIENTSECRET")); err != nil {
		clear(key.ClientSecret)
		return bitwarden.APIKey{}, err
	}
	ok := false
	defer func() {
		if !ok {
			clear(key.ClientSecret)
		}
	}()
	prompt := interactivePrompt(cmd, disabled)
	if prompt == nil && (key.ClientID == "" || len(key.ClientSecret) == 0) {
		return bitwarden.APIKey{}, errors.New("BWB_CLIENTID and BWB_CLIENTSECRET are required for non-interactive API key login")
	}
	if key.ClientID == "" {
		request.Label = "Client ID"
		id, err := prompt(cmd.Context(), request)
		if err != nil {
			clear(id)
			return bitwarden.APIKey{}, err
		}
		key.ClientID = string(id)
		clear(id)
	}
	if len(key.ClientSecret) == 0 {
		request.Label = "Client secret"
		var err error
		key.ClientSecret, err = prompt(cmd.Context(), request)
		if err != nil {
			return bitwarden.APIKey{}, err
		}
	}
	if err := key.Validate(); err != nil {
		return bitwarden.APIKey{}, err
	}
	ok = true
	return key, nil
}
