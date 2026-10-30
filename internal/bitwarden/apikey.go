package bitwarden

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/awnumar/memguard"
)

type APIKey struct {
	ClientID     string `json:"clientId"`
	ClientSecret []byte `json:"clientSecret"`
}

func (k APIKey) Validate() error {
	id, ok := strings.CutPrefix(k.ClientID, "user.")
	if !ok || !validID(id) {
		return errors.New("personal API key client_id must be user.<account UUID>")
	}
	if len(k.ClientSecret) == 0 || len(k.ClientSecret) > 4096 {
		return errors.New("invalid API key client_secret length")
	}
	return nil
}

func (c *Client) apiToken(ctx context.Context, apiKey APIKey) (tokens, error) {
	var response tokens
	if err := apiKey.Validate(); err != nil {
		return response, err
	}
	form := url.Values{"grant_type": {"client_credentials"}, "scope": {"api"},
		"client_id": {apiKey.ClientID}, "client_secret": {string(apiKey.ClientSecret)},
		"deviceType": {cliDeviceType()}, "deviceIdentifier": {c.deviceID}, "deviceName": {"bwb"}}
	body := []byte(form.Encode())
	defer clear(body)
	if err := c.request(ctx, "POST", c.identity+"/connect/token", body, true, "", &response); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == 400 || apiErr.Status == 401) {
			return tokens{}, ErrAuthentication
		}
		return tokens{}, err
	}
	if response.Access == "" {
		return tokens{}, errors.New("incomplete API key login response")
	}
	if response.APIUseKeyConnector || response.UserDecryptionOptions != nil &&
		response.UserDecryptionOptions.HasMasterPassword != nil && !*response.UserDecryptionOptions.HasMasterPassword {
		return tokens{}, errors.New("API key login requires a master password; Key Connector and passwordless accounts are unsupported")
	}
	if response.KDF == nil {
		return tokens{}, fmt.Errorf("API key token response is missing Kdf: %w", ErrUnsupported)
	}
	if response.Key == "" {
		return tokens{}, fmt.Errorf("API key token response is missing the protected vault key: %w", ErrUnsupported)
	}
	if err := response.kdfConfig().validate(); err != nil {
		return tokens{}, err
	}
	// API key sessions renew using client credentials, without depending on a refresh token.
	response.Refresh = ""
	return response, nil
}

func (c *Client) LoginAPIKey(ctx context.Context, password []byte, apiKey APIKey) (*Session, error) {
	if len(password) == 0 || len(password) > 4096 {
		return nil, errors.New("invalid password length")
	}
	response, err := c.apiToken(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if c.email == "" {
		var profile struct{ ID, Email string }
		if err := c.request(ctx, "GET", c.api+"/accounts/profile", nil, false, response.Access, &profile); err != nil {
			return nil, fmt.Errorf("resolve API key account: %w", err)
		}
		if "user."+profile.ID != apiKey.ClientID {
			return nil, errors.New("API key account does not match server profile")
		}
		email, err := normalizeEmail(profile.Email)
		if err != nil {
			return nil, fmt.Errorf("API key account email: %w", err)
		}
		resolved := *c
		resolved.email = email
		c = &resolved
	}
	stretched, _, err := derive(ctx, c.email, password, response.kdfConfig())
	if err != nil {
		return nil, err
	}
	defer clear(stretched)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	session, err := c.openSession(stretched, response.kdfConfig(), response)
	if err != nil {
		return nil, err
	}
	session.accountID = strings.TrimPrefix(apiKey.ClientID, "user.")
	if err := session.SetAPIKey(apiKey); err != nil {
		session.Close()
		return nil, err
	}
	return session, nil
}

func (s *Session) HasCredentials() bool { return s.refresh != nil || s.apiSecret != nil }

func (s *Session) SetAPIKey(apiKey APIKey) error {
	if !s.active() || s.refresh != nil {
		return ErrAuthentication
	}
	if err := apiKey.Validate(); err != nil {
		return err
	}
	if apiKey.ClientID != "user."+s.accountID {
		return errors.New("API key account does not match session")
	}
	if s.apiSecret != nil {
		s.apiSecret.Destroy()
	}
	s.apiClientID = apiKey.ClientID
	s.apiSecret = memguard.NewBufferFromBytes(bytes.Clone(apiKey.ClientSecret))
	return nil
}

func (s *Session) refreshAPIKey(ctx context.Context) error {
	response, err := s.client.apiToken(ctx, APIKey{ClientID: s.apiClientID, ClientSecret: s.apiSecret.Bytes()})
	if err != nil {
		return err
	}
	if response.Key != s.protectedKey || response.kdfConfig() != s.kdf {
		return ErrKeyChanged
	}
	if s.access != nil {
		s.access.Destroy()
	}
	s.access = memguard.NewBufferFromBytes([]byte(response.Access))
	return nil
}
