package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/awnumar/memguard"
)

type SSOCode struct {
	Code        []byte `json:"code"`
	Verifier    []byte `json:"verifier"`
	RedirectURI string `json:"redirectUri"`
	Identifier  string `json:"identifier,omitempty"`
}

var ErrSSODiscoveryUnsupported = errors.New("server does not support SSO domain discovery")

func ValidateSSOIdentifier(identifier string) error {
	if len(identifier) > 256 || identifier != "" && strings.TrimSpace(identifier) == "" || strings.ContainsAny(identifier, "\r\n\x00") {
		return errors.New("invalid SSO organization identifier")
	}
	return nil
}

// DiscoverSSO returns verified-domain matches, not a complete account SSO policy.
func (c *Client) DiscoverSSO(ctx context.Context) ([]string, error) {
	if c.email == "" {
		return nil, errors.New("email is required for SSO discovery")
	}
	var response struct {
		Data *[]struct {
			Identifier string `json:"organizationIdentifier"`
		} `json:"data"`
	}
	body, _ := json.Marshal(map[string]string{"email": c.email})
	if err := c.request(ctx, "POST", c.api+"/organizations/domain/sso/verified", body, false, "", &response); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == 404 || apiErr.Status == 405 || apiErr.Status == 501) {
			return nil, ErrSSODiscoveryUnsupported
		}
		return nil, fmt.Errorf("SSO discovery failed: %w", err)
	}
	if response.Data == nil {
		return nil, errors.New("invalid SSO discovery response")
	}
	identifiers := []string{}
	for _, org := range *response.Data {
		if org.Identifier == "" || ValidateSSOIdentifier(org.Identifier) != nil {
			return nil, errors.New("invalid SSO discovery identifier")
		}
		// Vaultwarden returns this placeholder even when SSO is disabled.
		if org.Identifier == "00000000-01DC-01DC-01DC-000000000000" {
			var validation struct{ Token string }
			err := c.request(ctx, "GET", c.identity+"/sso/prevalidate?domainHint="+url.QueryEscape(org.Identifier), nil, false, "", &validation)
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.Status == 400 {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("SSO discovery validation failed: %w", err)
			}
			if validation.Token == "" {
				return nil, errors.New("invalid SSO discovery validation response")
			}
		}
		identifiers = append(identifiers, org.Identifier)
	}
	return identifiers, nil
}

func (s SSOCode) Validate() error {
	if err := ValidateSSOIdentifier(s.Identifier); err != nil {
		return err
	}
	if len(s.Code) == 0 || len(s.Code) > 4096 {
		return errors.New("invalid SSO authorization code")
	}
	if len(s.Verifier) < 43 || len(s.Verifier) > 128 {
		return errors.New("invalid SSO PKCE verifier")
	}
	for _, b := range s.Verifier {
		if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("-._~", rune(b))) {
			return errors.New("invalid SSO PKCE verifier")
		}
	}
	// Bitwarden registers these loopback callbacks for its CLI OAuth client.
	for port := 8065; port <= 8070; port++ {
		if s.RedirectURI == fmt.Sprintf("http://localhost:%d", port) {
			return nil
		}
	}
	return errors.New("invalid SSO redirect URI")
}

func (c *Client) LoginSSO(ctx context.Context, password []byte, code SSOCode) (*Session, error) {
	pending, err := c.ExchangeSSO(ctx, code)
	if err != nil {
		return nil, err
	}
	defer pending.Close()
	return pending.UnlockSSO(ctx, password)
}

// ExchangeSSO retains tokens in locked memory until the caller unlocks or closes it.
func (c *Client) ExchangeSSO(ctx context.Context, code SSOCode) (*Session, error) {
	if err := code.Validate(); err != nil {
		return nil, err
	}
	form := url.Values{"grant_type": {"authorization_code"}, "client_id": {"cli"}, "scope": {"api offline_access"},
		"code": {string(code.Code)}, "code_verifier": {string(code.Verifier)}, "redirect_uri": {code.RedirectURI},
		"deviceType": {cliDeviceType()}, "deviceIdentifier": {c.deviceID}, "deviceName": {"bwb"}}
	body := []byte(form.Encode())
	defer clear(body)
	var response tokens
	if err := c.request(ctx, "POST", c.identity+"/connect/token", body, true, "", &response); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == 400 || apiErr.Status == 401) {
			return nil, fmt.Errorf("SSO code rejected, expired, or additional Bitwarden two-step login required; retry login --sso: %w", ErrAuthentication)
		}
		return nil, err
	}
	if response.Access == "" || response.Refresh == "" {
		return nil, errors.New("incomplete SSO login response")
	}
	if response.APIUseKeyConnector || response.KeyConnectorURL != "" || response.UserDecryptionOptions != nil &&
		response.UserDecryptionOptions.HasMasterPassword != nil && !*response.UserDecryptionOptions.HasMasterPassword {
		return nil, errors.New("SSO login requires a master password; Key Connector and passwordless Trusted Device accounts are unsupported")
	}
	if response.KDF == nil {
		return nil, fmt.Errorf("SSO response is missing Kdf: %w", ErrUnsupported)
	}
	if response.Key == "" {
		return nil, fmt.Errorf("SSO response is missing the master-password-protected vault key; complete account setup in the web vault: %w", ErrUnsupported)
	}
	if err := response.kdfConfig().validate(); err != nil {
		return nil, err
	}
	var profile struct{ ID, Email string }
	if err := c.request(ctx, "GET", c.api+"/accounts/profile", nil, false, response.Access, &profile); err != nil {
		return nil, fmt.Errorf("resolve SSO account: %w", err)
	}
	email, err := normalizeEmail(profile.Email)
	if err != nil || !validID(profile.ID) {
		return nil, errors.New("invalid SSO account profile")
	}
	if c.email != "" && c.email != email {
		return nil, errors.New("SSO account email differs from requested profile")
	}
	resolved := *c
	resolved.email = email
	return &Session{client: &resolved, accountID: profile.ID, protectedKey: response.Key, kdf: response.kdfConfig(),
		oauthClientID: "cli", ssoIdentifier: code.Identifier, access: memguard.NewBufferFromBytes([]byte(response.Access)),
		refresh: memguard.NewBufferFromBytes([]byte(response.Refresh))}, nil
}

func (s *Session) Email() string { return s.client.email }

func (s *Session) UnlockSSO(ctx context.Context, password []byte) (*Session, error) {
	if s.oauthClientID != "cli" || s.key != nil || s.access == nil || !s.access.IsAlive() || s.refresh == nil || !s.refresh.IsAlive() {
		return nil, ErrAuthentication
	}
	if len(password) == 0 || len(password) > 4096 {
		return nil, errors.New("invalid password length")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stretched, _, err := derive(ctx, s.client.email, password, s.kdf)
	if err != nil {
		return nil, err
	}
	defer clear(stretched)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	session, err := s.client.openSession(stretched, s.kdf, tokens{Key: s.protectedKey, Access: string(s.access.Bytes()), Refresh: string(s.refresh.Bytes())})
	if err != nil {
		return nil, err
	}
	session.oauthClientID, session.accountID, session.ssoIdentifier = "cli", s.accountID, s.ssoIdentifier
	return session, nil
}
