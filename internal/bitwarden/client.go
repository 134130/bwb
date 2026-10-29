package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/awnumar/memguard"
)

var ErrKeyChanged = errors.New("vault key changed; login again")

var ErrAuthentication = errors.New("login required or credentials rejected")

func cliDeviceType() string {
	switch runtime.GOOS {
	case "windows":
		return "23"
	case "darwin":
		return "24"
	default:
		return "25"
	}
}

// Client talks to Password Manager endpoints, never the public organization API.
type Client struct {
	api, identity, email, deviceID, server string
	http                                   *http.Client
}

func (c *Client) Identity() (server, email string) { return c.server, c.email }

// An empty email is resolved during API key or SSO login; password login still requires it.
func NewClient(server, email, deviceID string) (*Client, error) {
	u, err := url.Parse(server)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" || u.ForceQuery {
		return nil, errors.New("invalid server URL")
	}
	local := u.Hostname() == "localhost"
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		local = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return nil, errors.New("HTTPS required except for loopback servers")
	}
	if email != "" {
		email, err = normalizeEmail(email)
		if err != nil {
			return nil, err
		}
	}
	if !validID(deviceID) {
		return nil, errors.New("invalid device ID")
	}
	base := strings.TrimRight(u.String(), "/")
	api, identity := base+"/api", base+"/identity"
	switch base {
	case "https://vault.bitwarden.com":
		api, identity = "https://api.bitwarden.com", "https://identity.bitwarden.com"
	case "https://vault.bitwarden.eu":
		api, identity = "https://api.bitwarden.eu", "https://identity.bitwarden.eu"
	}
	return &Client{api: api, identity: identity, email: email, deviceID: deviceID, server: base,
		http: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", errors.New("invalid email address")
	}
	return email, nil
}

type APIError struct{ Status int }

func (e *APIError) Error() string { return fmt.Sprintf("vault API returned HTTP %d", e.Status) }
func (e *APIError) Is(target error) bool {
	return target == ErrAuthentication && e.Status == http.StatusUnauthorized
}

// Session is owned by one caller; it is not safe for concurrent operations.
type Session struct {
	client                  *Client
	key, access, refresh    *memguard.LockedBuffer
	apiSecret               *memguard.LockedBuffer
	apiClientID             string
	oauthClientID           string
	ssoIdentifier           string
	protectedKey, accountID string
	iterations              int
}

func (s *Session) Close() {
	for _, buffer := range []*memguard.LockedBuffer{s.key, s.access, s.refresh, s.apiSecret} {
		if buffer != nil {
			buffer.Destroy()
		}
	}
	s.key, s.access, s.refresh = nil, nil, nil
	s.apiSecret, s.apiClientID = nil, ""
	s.protectedKey, s.accountID = "", ""
}

func (s *Session) active() bool { return s.key != nil && s.key.IsAlive() }

type tokens struct {
	Access                string `json:"access_token"`
	Refresh               string `json:"refresh_token"`
	Key                   string `json:"Key"`
	KDF                   *int   `json:"Kdf"`
	Iterations            int    `json:"KdfIterations"`
	APIUseKeyConnector    bool   `json:"ApiUseKeyConnector"`
	KeyConnectorURL       string `json:"KeyConnectorUrl"`
	UserDecryptionOptions *struct {
		HasMasterPassword *bool `json:"HasMasterPassword"`
	} `json:"UserDecryptionOptions"`
}

func (c *Client) Login(ctx context.Context, password []byte) (*Session, error) {
	if c.email == "" {
		return nil, errors.New("email is required for password login")
	}
	if len(password) == 0 || len(password) > 4096 {
		return nil, errors.New("invalid password length")
	}
	var kdf struct {
		Type       *int `json:"kdf"`
		Iterations int  `json:"kdfIterations"`
	}
	data, _ := json.Marshal(map[string]string{"email": c.email})
	if err := c.request(ctx, "POST", c.identity+"/accounts/prelogin", data, false, "", &kdf); err != nil {
		return nil, err
	}
	if kdf.Type == nil || *kdf.Type != 0 {
		return nil, ErrUnsupported
	}
	stretched, hash, err := derive(c.email, password, kdf.Iterations)
	if err != nil {
		return nil, err
	}
	defer clear(stretched)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	form := url.Values{"grant_type": {"password"}, "client_id": {"web"}, "scope": {"api offline_access"},
		"username": {c.email}, "password": {hash}, "deviceType": {cliDeviceType()}, "deviceIdentifier": {c.deviceID}, "deviceName": {"bwb"}}
	var response tokens
	if err := c.request(ctx, "POST", c.identity+"/connect/token", []byte(form.Encode()), true, "", &response); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == 400 || apiErr.Status == 401) {
			return nil, ErrAuthentication
		}
		return nil, err
	}
	if response.Access == "" || response.Refresh == "" || response.Key == "" {
		return nil, errors.New("incomplete login response")
	}
	return c.openSession(stretched, kdf.Iterations, response)
}

func (c *Client) openSession(stretched []byte, iterations int, response tokens) (*Session, error) {
	key, err := decrypt(response.Key, stretched)
	if err != nil {
		return nil, fmt.Errorf("decrypt protected vault key: %w", err)
	}
	if len(key) != 64 {
		clear(key)
		return nil, ErrCipher
	}
	defer clear(key)
	session := &Session{client: c, key: memguard.NewBufferFromBytes(key),
		access: memguard.NewBufferFromBytes([]byte(response.Access)), protectedKey: response.Key, iterations: iterations}
	if response.Refresh != "" {
		session.refresh = memguard.NewBufferFromBytes([]byte(response.Refresh))
	}
	return session, nil
}

func (s *Session) Refresh(ctx context.Context) error {
	if !s.active() {
		return ErrAuthentication
	}
	if s.apiSecret != nil {
		return s.refreshAPIKey(ctx)
	}
	if s.refresh == nil {
		return ErrAuthentication
	}
	clientID := s.oauthClientID
	if clientID == "" {
		clientID = "web"
	}
	form := url.Values{"grant_type": {"refresh_token"}, "client_id": {clientID}, "refresh_token": {string(s.refresh.Bytes())}}
	var response tokens
	if err := s.client.request(ctx, "POST", s.client.identity+"/connect/token", []byte(form.Encode()), true, "", &response); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == 400 || apiErr.Status == 401) {
			return ErrAuthentication
		}
		return err
	}
	if response.Access == "" {
		return errors.New("incomplete refresh response")
	}
	if s.access != nil {
		s.access.Destroy()
	}
	s.access = memguard.NewBufferFromBytes([]byte(response.Access))
	if response.Refresh != "" {
		s.refresh.Destroy()
		s.refresh = memguard.NewBufferFromBytes([]byte(response.Refresh))
	}
	return nil
}

const maxResponse = 32 << 20

func (c *Client) request(ctx context.Context, method, endpoint string, body []byte, form bool, token string, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("invalid API request")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Bitwarden-Client-Version", "2026.10.0")
	if body != nil {
		contentType := "application/json"
		if form {
			contentType = "application/x-www-form-urlencoded"
		}
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("vault API connection failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &APIError{Status: response.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponse+1))
	if err != nil {
		return errors.New("cannot read vault API response")
	}
	defer clear(data)
	if len(data) > maxResponse {
		return errors.New("vault API response exceeds 32 MiB")
	}
	if !utf8.Valid(data) || json.Unmarshal(data, result) != nil {
		return errors.New("invalid vault API JSON response")
	}
	return nil
}
