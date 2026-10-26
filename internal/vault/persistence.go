package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/go-daemonkit/storage"
	"github.com/google/uuid"
)

// Credentials is the native credential-store boundary; tests inject I/O failures here.
// Production always uses daemonkit Secrets, with no file or memory fallback.
type Credentials interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte) error
	Delete(context.Context, string) error
}

type record struct {
	Version    int              `json:"version"`
	AuthMethod string           `json:"authMethod,omitempty"`
	Profile    string           `json:"profile"`
	Cache      *bitwarden.Cache `json:"cache,omitempty"`
	Credential string           `json:"credential,omitempty"`
	Garbage    []string         `json:"garbage,omitempty"`
	Deleting   bool             `json:"deleting,omitempty"`
}

type credential struct {
	Server, Email, AccountID string
	Token                    []byte
	ClientID                 string `json:",omitempty"`
	ClientSecret             []byte `json:",omitempty"`
}

func (c *credential) clear() {
	clear(c.Token)
	clear(c.ClientSecret)
}

const chunkSize = 512 << 10
const maxChunks = 128

func count(tx *storage.ReadTx, name string) (int, error) {
	data, err := tx.Get(name + "/count")
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var n int
	if json.Unmarshal(data, &n) != nil || n < 1 || n > maxChunks {
		return 0, errors.New("invalid cache chunk count")
	}
	return n, nil
}

func chunkKey(name string, i int) string { return fmt.Sprintf("%s/chunk/%03d", name, i) }

func (s *Service) load(ctx context.Context, name string) (*record, error) {
	if s.Store == nil {
		return nil, errors.New("vault storage unavailable")
	}
	var data []byte
	err := s.Store.View(ctx, func(tx *storage.ReadTx) error {
		n, err := count(tx, name)
		if err != nil {
			return err
		}
		for i := range n {
			part, err := tx.Get(chunkKey(name, i))
			if err != nil {
				return err
			}
			if len(part) == 0 || len(part) > chunkSize || (i < n-1 && len(part) != chunkSize) {
				return errors.New("invalid cache chunk size")
			}
			data = append(data, part...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := &record{Version: 1, Profile: name}
	if len(data) == 0 {
		return r, nil
	}
	if json.Unmarshal(data, r) != nil || r.Profile != name || len(r.Garbage) > 4 {
		return nil, errors.New("invalid profile cache")
	}
	if !(r.Version == 1 && r.AuthMethod == "" || r.Version == 2 && r.AuthMethod == "apikey") {
		return nil, errors.New("unsupported profile authentication version")
	}
	if r.Cache != nil && r.Cache.Version != 1 {
		return nil, errors.New("unsupported cache version")
	}
	valid := func(key string) bool {
		suffix, ok := strings.CutPrefix(key, name+"/")
		id, err := uuid.Parse(suffix)
		return ok && err == nil && id != uuid.Nil && id.String() == suffix
	}
	if (r.Cache == nil) != (r.Credential == "") || (r.Credential != "" && !valid(r.Credential)) {
		return nil, errors.New("invalid credential binding")
	}
	for _, key := range r.Garbage {
		if !valid(key) || key == r.Credential {
			return nil, errors.New("invalid credential cleanup journal")
		}
	}
	return r, nil
}

// put atomically replaces all chunks; daemonkit limits each value to 1 MiB.
func (s *Service) put(ctx context.Context, name string, r *record) error {
	var data []byte
	var err error
	if r != nil {
		data, err = json.Marshal(r)
		if err != nil {
			return err
		}
	}
	n := (len(data) + chunkSize - 1) / chunkSize
	if n > maxChunks {
		return errors.New("vault cache exceeds 64 MiB")
	}
	return s.Store.Update(ctx, func(tx *storage.WriteTx) error {
		old, err := count(&tx.ReadTx, name)
		if err != nil {
			return err
		}
		for i := range n {
			if err := tx.Set(chunkKey(name, i), data[i*chunkSize:min((i+1)*chunkSize, len(data))]); err != nil {
				return err
			}
		}
		for i := n; i < old; i++ {
			if err := tx.Delete(chunkKey(name, i)); err != nil {
				return err
			}
		}
		if n == 0 {
			return tx.Delete(name + "/count")
		}
		encoded, _ := json.Marshal(n)
		return tx.Set(name+"/count", encoded)
	})
}

func (s *Service) cleanup(ctx context.Context, r *record) error {
	if len(r.Garbage) == 0 {
		return nil
	}
	if s.Secrets == nil {
		return errors.New("credential store unavailable")
	}
	for _, key := range r.Garbage {
		if err := s.Secrets.Delete(ctx, key); err != nil {
			return err
		}
	}
	r.Garbage = nil
	return s.put(ctx, r.Profile, r)
}

func (s *Service) saveLogin(ctx context.Context, r *record, cache *bitwarden.Cache, auth credential) error {
	if s.Secrets == nil {
		return errors.New("credential store unavailable")
	}
	if err := s.cleanup(ctx, r); err != nil {
		return err
	}
	key := r.Profile + "/" + uuid.NewString()
	r.Garbage = append(r.Garbage, key)
	// Persist the intent before the nontransactional OS write, including uncertain writes.
	if err := s.put(ctx, r.Profile, r); err != nil {
		return err
	}
	auth.Server, auth.Email, auth.AccountID = cache.Server, cache.Email, cache.AccountID()
	data, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	defer clear(data)
	if err := s.Secrets.Set(ctx, key, data); err != nil {
		return err
	}
	old := r.Credential
	r.Cache, r.Credential, r.Garbage = cache, key, nil
	r.Version, r.AuthMethod = 1, ""
	if auth.ClientID != "" {
		r.Version, r.AuthMethod = 2, "apikey"
	}
	if old != "" {
		r.Garbage = []string{old}
	}
	// Old credentials remain journaled until the next mutation or logout.
	return s.put(ctx, r.Profile, r)
}

func (s *Service) loadCredential(ctx context.Context, r *record) (credential, error) {
	if s.Secrets == nil {
		return credential{}, errors.New("credential store unavailable")
	}
	data, err := s.Secrets.Get(ctx, r.Credential)
	if err != nil {
		return credential{}, err
	}
	defer clear(data)
	var c credential
	if json.Unmarshal(data, &c) != nil {
		c.clear()
		return credential{}, errors.New("invalid saved credential")
	}
	valid := len(c.Token) != 0 && c.ClientID == "" && len(c.ClientSecret) == 0
	if r.AuthMethod == "apikey" {
		key := bitwarden.APIKey{ClientID: c.ClientID, ClientSecret: c.ClientSecret}
		valid = len(c.Token) == 0 && key.Validate() == nil && c.ClientID == "user."+c.AccountID
	}
	if !valid || c.Server != r.Cache.Server || c.Email != r.Cache.Email || c.AccountID != r.Cache.AccountID() {
		c.clear()
		return credential{}, errors.New("saved credential identity mismatch")
	}
	return c, nil
}

func (s *Service) saveToken(ctx context.Context, r *record, token []byte) error {
	if r.AuthMethod != "" || len(token) == 0 {
		return errors.New("invalid refresh token credential")
	}
	if s.Secrets == nil {
		return errors.New("credential store unavailable")
	}
	data, err := json.Marshal(credential{Server: r.Cache.Server, Email: r.Cache.Email, AccountID: r.Cache.AccountID(), Token: token})
	if err != nil {
		return err
	}
	defer clear(data)
	return s.Secrets.Set(ctx, r.Credential, data)
}
