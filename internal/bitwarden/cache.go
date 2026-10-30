package bitwarden

import (
	"bytes"
	"context"
	"errors"

	"github.com/awnumar/memguard"
)

// Cache contains server ciphertext and public derivation metadata, never tokens or plaintext keys.
type Cache struct {
	OAuthClientID string       `json:"oauthClientId,omitempty"`
	SSOIdentifier string       `json:"ssoIdentifier,omitempty"`
	Version       int          `json:"version"`
	Server        string       `json:"server"`
	Email         string       `json:"email"`
	DeviceID      string       `json:"deviceId"`
	KDF           int          `json:"kdf"`
	Iterations    int          `json:"iterations"`
	Memory        int          `json:"memory,omitempty"`
	Parallelism   int          `json:"parallelism,omitempty"`
	Snapshot      syncResponse `json:"snapshot"`
}

func (c *Client) Matches(cache *Cache) bool {
	return c.server == cache.Server && c.email == cache.Email
}

func (s *Session) Cache(snapshot *Snapshot) *Cache {
	return &Cache{Version: 1, Server: s.client.server, Email: s.client.email, DeviceID: s.client.deviceID,
		OAuthClientID: s.oauthClientID, SSOIdentifier: s.ssoIdentifier, KDF: s.kdf.Type, Iterations: s.kdf.Iterations, Memory: s.kdf.Memory, Parallelism: s.kdf.Parallelism, Snapshot: snapshot.response}
}

// Unlock validates cached identity and readable contents without accessing the server or credential store.
func Unlock(ctx context.Context, cache *Cache, password []byte) (*Session, *Snapshot, error) {
	if cache == nil || cache.Version != 1 || cache.Email == "" {
		return nil, nil, ErrUnsupported
	}
	if cache.OAuthClientID != "" && cache.OAuthClientID != "cli" {
		return nil, nil, ErrUnsupported
	}
	if len(password) == 0 || len(password) > 4096 {
		return nil, nil, errors.New("invalid password length")
	}
	client, err := NewClient(cache.Server, cache.Email, cache.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if client.email != cache.Email || client.server != cache.Server {
		return nil, nil, errors.New("noncanonical cache identity")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	config := kdfConfig{cache.KDF, cache.Iterations, cache.Memory, cache.Parallelism}
	stretched, _, err := derive(ctx, cache.Email, password, config)
	if err != nil {
		return nil, nil, err
	}
	defer clear(stretched)
	key, err := decrypt(cache.Snapshot.Profile.Key, stretched)
	if err != nil {
		return nil, nil, ErrCipher
	}
	defer clear(key)
	if len(key) != 64 {
		return nil, nil, ErrCipher
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	session := &Session{client: client, key: memguard.NewBufferFromBytes(key),
		oauthClientID: cache.OAuthClientID, ssoIdentifier: cache.SSOIdentifier, protectedKey: cache.Snapshot.Profile.Key, kdf: config}
	snapshot, _, err := session.validateSnapshot(ctx, cache.Snapshot)
	if err != nil {
		session.Close()
		return nil, nil, err
	}
	return session, snapshot, nil
}

func (c *Cache) AccountID() string { return c.Snapshot.Profile.ID }

// RefreshToken returns an owned copy; callers must clear it after credential-store I/O.
func (s *Session) RefreshToken() []byte {
	if s.refresh == nil {
		return nil
	}
	return bytes.Clone(s.refresh.Bytes())
}

func (s *Session) SetRefreshToken(token []byte) error {
	if !s.active() {
		return ErrAuthentication
	}
	if len(token) == 0 {
		return ErrAuthentication
	}
	if s.refresh != nil {
		s.refresh.Destroy()
	}
	s.refresh = memguard.NewBufferFromBytes(bytes.Clone(token))
	return nil
}
