package bitwarden

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"strings"

	"github.com/awnumar/memguard"
)

// Organization keys are encrypted to the account's RSA public key.
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type Collection struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	ReadOnly       bool   `json:"readOnly"`
	HidePasswords  bool   `json:"hidePasswords"`
}

type organizationKeys map[string]*memguard.LockedBuffer

func (keys organizationKeys) close() {
	for _, key := range keys {
		key.Destroy()
	}
}

func (s *Session) organizationKeys(ctx context.Context, profile Profile) (organizationKeys, error) {
	keys := organizationKeys{}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(profile.Organizations) == 0 {
		return keys, nil
	}
	encoded, err := decrypt(profile.PrivateKey, s.key.Bytes())
	if err != nil {
		return nil, err
	}
	defer clear(encoded)
	if len(encoded) > 16<<10 {
		return nil, ErrUnsupported
	}
	parsed, err := x509.ParsePKCS8PrivateKey(encoded)
	if err != nil {
		return nil, ErrCipher
	}
	private, ok := parsed.(*rsa.PrivateKey)
	if !ok || private.N.BitLen() < 2048 || private.N.BitLen() > 4096 {
		return nil, ErrUnsupported
	}
	// Parsed RSA big integers are temporary Go allocations, never retained by the session.
	for _, org := range profile.Organizations {
		if err := ctx.Err(); err != nil {
			keys.close()
			return nil, err
		}
		if !validID(org.ID) || keys[org.ID] != nil {
			keys.close()
			return nil, errors.New("invalid or duplicate organization ID")
		}
		key, err := decryptOrganizationKey(org.Key, private)
		if err != nil {
			keys.close()
			return nil, err
		}
		keys[org.ID] = memguard.NewBufferFromBytes(key)
	}
	return keys, nil
}

func decryptOrganizationKey(value string, private *rsa.PrivateKey) ([]byte, error) {
	kind, encoded, ok := strings.Cut(value, ".")
	if !ok {
		return nil, ErrCipher
	}
	var digest hash.Hash
	switch kind {
	case "3":
		digest = sha256.New()
	case "4":
		digest = sha1.New() // Bitwarden's legacy RSA-OAEP-SHA1 envelope.
	default:
		return nil, ErrUnsupported
	}
	ciphertext, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(ciphertext) != private.Size() {
		return nil, ErrCipher
	}
	key, err := rsa.DecryptOAEP(digest, rand.Reader, private, ciphertext, nil)
	if err != nil || len(key) != 64 {
		clear(key)
		return nil, ErrCipher
	}
	return key, nil
}

func decryptCollections(encrypted []Collection, keys organizationKeys) ([]Collection, error) {
	collections := make([]Collection, 0, len(encrypted))
	seen := map[string]bool{}
	for _, collection := range encrypted {
		key := keys[collection.OrganizationID]
		if !validID(collection.ID) || seen[collection.ID] || key == nil {
			return nil, errors.New("invalid collection relationship")
		}
		seen[collection.ID] = true
		name, err := decryptText(collection.Name, key.Bytes())
		if err != nil {
			return nil, err
		}
		collection.Name = name
		collections = append(collections, collection)
	}
	return collections, nil
}

func validateCollections(item Item, collections []Collection) error {
	seen := map[string]bool{}
	for _, id := range item.CollectionIDs {
		if !validID(id) || seen[id] || item.OrganizationID == nil {
			return errors.New("invalid item collection relationship")
		}
		seen[id] = true
		found := false
		for _, collection := range collections {
			if collection.ID == id && collection.OrganizationID == *item.OrganizationID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("unknown item collection")
		}
	}
	return nil
}

func (s *Session) decryptOwnedItem(raw []byte, keys organizationKeys) (Item, error) {
	var envelope struct {
		OrganizationID *string `json:"organizationId"`
		ViewPassword   *bool   `json:"viewPassword"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return Item{}, errors.New("invalid item")
	}
	key := s.key
	if envelope.OrganizationID != nil {
		key = keys[*envelope.OrganizationID]
		if key == nil {
			return Item{}, errors.New("unknown item organization")
		}
		if envelope.ViewPassword == nil {
			return Item{}, errors.New("missing organization item permissions")
		}
	}
	return decryptItem(raw, key.Bytes())
}

var ErrPermission = errors.New("item field access restricted by organization policy")
