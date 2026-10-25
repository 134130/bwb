package bitwarden

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"testing"

	"github.com/awnumar/memguard"
)

func TestOrganizationRSAEnvelopes(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	key := bytes.Repeat([]byte{19}, 64)
	for _, tc := range []struct {
		prefix string
		digest func() hash.Hash
	}{{"3.", sha256.New}, {"4.", sha1.New}} {
		ciphertext, err := rsa.EncryptOAEP(tc.digest(), rand.Reader, &private.PublicKey, key, nil)
		if err != nil {
			t.Fatal(err)
		}
		envelope := tc.prefix + base64.StdEncoding.EncodeToString(ciphertext)
		got, err := decryptOrganizationKey(envelope, private)
		if err != nil || !bytes.Equal(got, key) {
			t.Fatal("RSA envelope failed", err)
		}
		clear(got)
		ciphertext[0] ^= 1
		if got, err := decryptOrganizationKey(tc.prefix+base64.StdEncoding.EncodeToString(ciphertext), private); err == nil || got != nil {
			t.Fatal("tampered RSA accepted")
		}
	}
	for _, value := range []string{"", "4.invalid", "4.", "5.invalid", "3.a|b"} {
		if got, err := decryptOrganizationKey(value, private); err == nil || got != nil {
			t.Fatal("invalid RSA accepted")
		}
	}
}

func TestOwnedItemKeyAndRestrictions(t *testing.T) {
	const id = "20000000-0000-4000-8000-000000000001"
	const oid = "30000000-0000-4000-8000-000000000001"
	userKey, orgKey, itemKey := bytes.Repeat([]byte{1}, 64), bytes.Repeat([]byte{2}, 64), bytes.Repeat([]byte{3}, 64)
	encrypt := func(value []byte, key []byte) string {
		n := 16 - len(value)%16
		padded := append(bytes.Clone(value), bytes.Repeat([]byte{byte(n)}, n)...)
		return encryptPadded(t, padded, key)
	}
	session := &Session{key: memguard.NewBufferFromBytes(bytes.Clone(userKey))}
	defer session.Close()
	keys := organizationKeys{oid: memguard.NewBufferFromBytes(bytes.Clone(orgKey))}
	defer keys.close()
	payload := map[string]any{"id": id, "type": 1, "organizationId": oid, "viewPassword": false,
		"name": encrypt([]byte("name"), itemKey), "key": encrypt(itemKey, orgKey),
		"login": map[string]any{"password": encrypt([]byte("secret"), itemKey)}}
	decode := func() (Item, error) { raw, _ := json.Marshal(payload); return session.decryptOwnedItem(raw, keys) }
	item, err := decode()
	if err != nil {
		t.Fatal(err)
	}
	vault := &Vault{Items: []Item{item}}
	if value, err := vault.Read(id, "name"); err != nil || value != "name" {
		t.Fatal("restricted name unavailable", err)
	}
	for _, field := range []string{"login/password", "login/username", "notes", "fields/token"} {
		if value, err := vault.Read(id, field); !errors.Is(err, ErrPermission) || value != "" {
			t.Fatal("restriction bypass", field, err)
		}
	}
	payload["viewPassword"] = true
	item, err = decode()
	if err != nil {
		t.Fatal(err)
	}
	if value, err := (&Vault{Items: []Item{item}}).Read(id, "login/password"); err != nil || value != "secret" {
		t.Fatal("item key failed", err)
	}
	delete(payload, "viewPassword")
	if _, err := decode(); err == nil {
		t.Fatal("missing policy accepted")
	}
	payload["viewPassword"] = true
	payload["key"] = encrypt(itemKey, userKey)
	if _, err := decode(); err == nil {
		t.Fatal("user key fallback for organization item")
	}
	payload["organizationId"] = "unknown"
	if _, err := decode(); err == nil {
		t.Fatal("unknown organization accepted")
	}
	// Optional item keys also apply to personal items.
	delete(payload, "organizationId")
	if _, err := decode(); err != nil {
		t.Fatal("personal item key failed", err)
	}
	payload["key"] = encrypt([]byte("short key"), userKey)
	if _, err := decode(); err == nil {
		t.Fatal("invalid item key size accepted")
	}
}

func TestCollectionRelationships(t *testing.T) {
	const first = "10000000-0000-4000-8000-000000000001"
	const second = "10000000-0000-4000-8000-000000000002"
	org, other := "org", "other"
	collections := []Collection{{ID: first, OrganizationID: org}, {ID: second, OrganizationID: org}}
	item := Item{OrganizationID: &org, CollectionIDs: []string{first, second}, FolderID: new(first)}
	if err := validateCollections(item, collections); err != nil {
		t.Fatal("multiple collections and folder rejected", err)
	}
	for _, ids := range [][]string{{first, first}, {"unknown"}} {
		item.CollectionIDs = ids
		if err := validateCollections(item, collections); err == nil {
			t.Fatal("invalid collection relationship accepted")
		}
	}
	item.CollectionIDs = []string{first}
	for _, owner := range []*string{nil, &other} {
		item.OrganizationID = owner
		if err := validateCollections(item, collections); err == nil {
			t.Fatal("cross-organization collection accepted")
		}
	}
}
