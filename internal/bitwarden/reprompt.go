package bitwarden

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"maps"
)

// VerifyPassword checks the current protected vault key without replacing the session.
func (s *Session) VerifyPassword(ctx context.Context, password []byte) error {
	if !s.active() {
		return ErrAuthentication
	}
	if len(password) == 0 || len(password) > 4096 {
		return errors.New("invalid password length")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	stretched, _, err := derive(s.client.email, password, s.iterations)
	if err != nil {
		return err
	}
	defer clear(stretched)
	key, err := decrypt(s.protectedKey, stretched)
	if err != nil {
		return errors.New("incorrect master password")
	}
	defer clear(key)
	if subtle.ConstantTimeCompare(key, s.key.Bytes()) != 1 {
		return errors.New("incorrect master password")
	}
	return ctx.Err()
}

func (s *Snapshot) RequiresReprompt(id string) (bool, error) {
	raw, ok := s.items[id]
	if !ok {
		return false, ErrNotFound
	}
	var item struct {
		Reprompt int `json:"reprompt"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return false, err
	}
	if item.Reprompt != 0 && item.Reprompt != 1 {
		return false, ErrUnsupported
	}
	return item.Reprompt == 1, nil
}

// AllowReprompt is for daemon-authorized reads only. Never changes the cached snapshot or organization permissions.
func (s *Snapshot) AllowReprompt(id string) (*Snapshot, error) {
	raw, ok := s.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	envelope["reprompt"] = json.RawMessage("0")
	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	copy := *s
	copy.items = maps.Clone(s.items)
	copy.items[id] = raw
	return &copy, nil
}
