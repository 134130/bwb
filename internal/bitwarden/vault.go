package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("item or field not found")

var ErrReprompt = errors.New("item requires master password reprompt; open it in the Bitwarden app")

var payloadNames = map[int]string{1: "login", 2: "secureNote", 3: "card", 4: "identity", 5: "sshKey", 6: "bankAccount", 7: "driversLicense", 8: "passport"}

var encryptedFields = map[int][]string{
	1: {"username", "password", "totp"},
	2: {},
	3: {"cardholderName", "brand", "number", "expMonth", "expYear", "code"},
	4: {"title", "firstName", "middleName", "lastName", "address1", "address2", "address3", "city", "state", "postalCode", "country", "company", "email", "phone", "ssn", "username", "passportNumber", "licenseNumber"},
	5: {"privateKey", "publicKey", "keyFingerprint"},
	6: {"bankName", "nameOnAccount", "accountType", "accountNumber", "routingNumber", "branchNumber", "pin", "swiftCode", "iban", "bankContactPhone"},
	7: {"firstName", "middleName", "lastName", "dateOfBirth", "licenseNumber", "issuingCountry", "issuingState", "issueDate", "expirationDate", "issuingAuthority", "licenseClass"},
	8: {"surname", "givenName", "dateOfBirth", "sex", "birthPlace", "nationality", "issuingCountry", "passportNumber", "passportType", "nationalIdentificationNumber", "issuingAuthority", "issueDate", "expirationDate"},
}

type Field struct {
	Name     *string `json:"name"`
	Value    *string `json:"value"`
	Type     int     `json:"type"`
	LinkedID *int    `json:"linkedId,omitempty"`
}

type Item struct {
	ID             string   `json:"id"`
	Type           int      `json:"type"`
	OrganizationID *string  `json:"organizationId"`
	CollectionIDs  []string `json:"collectionIds"`
	ViewPassword   *bool    `json:"viewPassword,omitempty"`
	Edit           *bool    `json:"edit,omitempty"`
	FolderID       *string  `json:"folderId"`
	Name           string   `json:"name"`
	Notes          *string  `json:"notes"`
	Favorite       bool     `json:"favorite"`
	Reprompt       int      `json:"reprompt"`
	Fields         []Field  `json:"fields"`
	// Data contains the selected type's payload. Only known encrypted fields are decrypted.
	Data map[string]any `json:"data"`
}

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Profile struct {
	ID            string         `json:"id"`
	Email         string         `json:"email"`
	Key           string         `json:"key"`
	PrivateKey    string         `json:"privateKey,omitempty"`
	Organizations []Organization `json:"organizations,omitempty"`
}
type Vault struct {
	Profile     Profile
	Folders     []Folder
	Collections []Collection
	Items       []Item
}

type syncResponse struct {
	Profile     Profile           `json:"profile"`
	Folders     []Folder          `json:"folders"`
	Ciphers     []json.RawMessage `json:"ciphers"`
	Collections []Collection      `json:"collections,omitempty"`
}

func validID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

// Snapshot holds only server ciphertext and metadata; reads never contact the server.
type Snapshot struct {
	profile  Profile
	response syncResponse
	items    map[string]json.RawMessage
}

func (s *Session) Sync(ctx context.Context) (*Vault, error) {
	_, vault, err := s.sync(ctx)
	return vault, err
}

func (s *Session) SyncSnapshot(ctx context.Context) (*Snapshot, error) {
	snapshot, _, err := s.sync(ctx)
	return snapshot, err
}

func (s *Session) sync(ctx context.Context) (*Snapshot, *Vault, error) {
	if !s.active() {
		return nil, nil, ErrAuthentication
	}
	if s.access == nil {
		if err := s.Refresh(ctx); err != nil {
			return nil, nil, err
		}
	}
	var response syncResponse
	err := s.client.request(ctx, "GET", s.client.api+"/sync", nil, false, string(s.access.Bytes()), &response)
	if errors.Is(err, ErrAuthentication) {
		if err := s.Refresh(ctx); err != nil {
			return nil, nil, err
		}
		err = s.client.request(ctx, "GET", s.client.api+"/sync", nil, false, string(s.access.Bytes()), &response)
	}
	if err != nil {
		return nil, nil, err
	}
	return s.validateSnapshot(ctx, response)
}

func (s *Session) validateSnapshot(ctx context.Context, response syncResponse) (*Snapshot, *Vault, error) {
	if !validID(response.Profile.ID) || strings.ToLower(strings.TrimSpace(response.Profile.Email)) != s.client.email ||
		(s.accountID != "" && s.accountID != response.Profile.ID) {
		return nil, nil, errors.New("vault account does not match session")
	}
	if response.Profile.Key != s.protectedKey {
		return nil, nil, ErrKeyChanged
	}
	if response.Folders == nil || response.Ciphers == nil {
		return nil, nil, errors.New("incomplete sync response")
	}
	vault, err := s.decryptSnapshot(ctx, response)
	if err != nil {
		return nil, nil, err
	}
	s.accountID = response.Profile.ID
	items := make(map[string]json.RawMessage, len(vault.Items))
	for i, item := range vault.Items {
		items[item.ID] = response.Ciphers[i]
	}
	return &Snapshot{profile: response.Profile, response: response, items: items}, vault, nil
}

func (s *Session) decryptSnapshot(ctx context.Context, response syncResponse) (*Vault, error) {
	keys, err := s.organizationKeys(ctx, response.Profile)
	if err != nil {
		return nil, fmt.Errorf("decrypt organization keys: %w", err)
	}
	defer keys.close()
	collections, err := decryptCollections(response.Collections, keys)
	if err != nil {
		return nil, fmt.Errorf("decrypt collections: %w", err)
	}
	vault := &Vault{Collections: collections, Profile: response.Profile, Folders: make([]Folder, 0, len(response.Folders)), Items: make([]Item, 0, len(response.Ciphers))}
	folders := make(map[string]bool)
	for _, folder := range response.Folders {
		if !validID(folder.ID) || folders[folder.ID] {
			return nil, errors.New("invalid or duplicate folder ID")
		}
		folders[folder.ID] = true
		name, err := decryptText(folder.Name, s.key.Bytes())
		if err != nil {
			return nil, fmt.Errorf("decrypt folder name: %w", err)
		}
		vault.Folders = append(vault.Folders, Folder{ID: folder.ID, Name: name})
	}
	ids := make(map[string]bool)
	for i, raw := range response.Ciphers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		item, err := s.decryptOwnedItem(raw, keys)
		if err != nil {
			return nil, fmt.Errorf("decrypt vault item %d: %w", i+1, err)
		}
		if err := validateCollections(item, collections); err != nil {
			return nil, err
		}
		if ids[item.ID] {
			return nil, errors.New("duplicate item ID")
		}
		ids[item.ID] = true
		if item.FolderID != nil && !folders[*item.FolderID] {
			return nil, errors.New("unknown folder ID")
		}
		vault.Items = append(vault.Items, item)
	}
	return vault, nil
}

func decryptText(value string, key []byte) (string, error) {
	plain, err := decrypt(value, key)
	if err != nil {
		return "", err
	}
	defer clear(plain)
	if !utf8.Valid(plain) {
		return "", errors.New("invalid UTF-8 in vault field")
	}
	return string(plain), nil
}

func decryptOptional(value **string, key []byte) error {
	if *value == nil {
		return nil
	}
	plain, err := decryptText(**value, key)
	if err != nil {
		return err
	}
	*value = &plain
	return nil
}

func decryptMap(data map[string]any, names []string, key []byte) error {
	for _, name := range names {
		value, ok := data[name]
		if !ok || value == nil {
			continue
		}
		encrypted, ok := value.(string)
		if !ok {
			return errors.New("invalid encrypted field type")
		}
		plain, err := decryptText(encrypted, key)
		if err != nil {
			return err
		}
		data[name] = plain
	}
	return nil
}

func decryptItem(raw []byte, userKey []byte) (Item, error) {
	var item Item
	var envelope map[string]json.RawMessage
	if json.Unmarshal(raw, &item) != nil || json.Unmarshal(raw, &envelope) != nil {
		return Item{}, errors.New("invalid item")
	}
	if !validID(item.ID) {
		return Item{}, errors.New("invalid item ID")
	}
	payload, supported := payloadNames[item.Type]
	if !supported {
		return Item{}, fmt.Errorf("item type %d: %w", item.Type, ErrUnsupported)
	}
	// Reprompt requires a separate authorization flow.
	var itemKey *string
	if json.Unmarshal(envelope["key"], &itemKey) != nil && envelope["key"] != nil {
		return Item{}, errors.New("invalid item key")
	}
	if item.Reprompt != 0 && item.Reprompt != 1 {
		return Item{}, fmt.Errorf("reprompt type %d: %w", item.Reprompt, ErrUnsupported)
	}
	key, err := decryptItemKey(itemKey, userKey)
	if err != nil {
		return Item{}, err
	}
	defer clear(key)
	userKey = key
	if json.Unmarshal(envelope[payload], &item.Data) != nil || item.Data == nil {
		return Item{}, errors.New("missing item payload")
	}
	item.Name, err = decryptText(item.Name, userKey)
	if err != nil {
		return Item{}, err
	}
	if item.Reprompt == 1 {
		// shortcut: expose only names until per-item master password verification is implemented.
		item.Notes, item.Fields, item.Data = nil, nil, nil
		return item, nil
	}
	if err := decryptOptional(&item.Notes, userKey); err != nil {
		return Item{}, err
	}
	for i := range item.Fields {
		field := &item.Fields[i]
		if field.Type < 0 || field.Type > 3 {
			return Item{}, fmt.Errorf("custom field type %d: %w", field.Type, ErrUnsupported)
		}
		if err := decryptOptional(&field.Name, userKey); err != nil {
			return Item{}, err
		}
		if err := decryptOptional(&field.Value, userKey); err != nil {
			return Item{}, err
		}
	}
	if err := decryptMap(item.Data, encryptedFields[item.Type], userKey); err != nil {
		return Item{}, err
	}
	if item.Type == 1 {
		if credentials, exists := item.Data["fido2Credentials"]; exists && credentials != nil {
			values, ok := credentials.([]any)
			if !ok || len(values) != 0 {
				return Item{}, fmt.Errorf("passkey credentials: %w", ErrUnsupported)
			}
		}
		if uris := item.Data["uris"]; uris != nil {
			values, ok := uris.([]any)
			if !ok {
				return Item{}, errors.New("invalid login URIs")
			}
			for _, value := range values {
				uri, ok := value.(map[string]any)
				if !ok {
					return Item{}, errors.New("invalid login URI")
				}
				if err := decryptMap(uri, []string{"uri", "uriChecksum"}, userKey); err != nil {
					return Item{}, err
				}
			}
		}
	}
	return item, nil
}

// Read returns an explicitly requested plaintext field, including empty values.
func (v *Vault) Read(id, path string) (string, error) {
	if !validID(id) {
		return "", errors.New("invalid item ID")
	}
	for _, item := range v.Items {
		if item.ID != id {
			continue
		}
		if item.Reprompt != 0 && path != "name" {
			return "", ErrReprompt
		}
		// shortcut: restricted items expose only their name until field-level policy parity is verified.
		if item.ViewPassword != nil && !*item.ViewPassword && path != "name" {
			return "", ErrPermission
		}
		switch path {
		case "name":
			return item.Name, nil
		case "notes":
			if item.Notes != nil {
				return *item.Notes, nil
			}
			return "", ErrNotFound
		}
		kind, fieldName, found := strings.Cut(path, "/")
		if !found || fieldName == "" {
			return "", errors.New("expected TYPE/FIELD, fields/NAME, name, or notes")
		}
		if kind == "fields" {
			var matches []*string
			for _, field := range item.Fields {
				if field.Name != nil && *field.Name == fieldName {
					matches = append(matches, field.Value)
				}
			}
			if len(matches) > 1 {
				return "", errors.New("ambiguous custom field name")
			}
			if len(matches) == 1 && matches[0] != nil {
				return *matches[0], nil
			}
			return "", ErrNotFound
		}
		if kind != payloadNames[item.Type] {
			return "", ErrNotFound
		}
		if kind == "login" && strings.HasPrefix(fieldName, "uris/") {
			index, err := strconv.Atoi(strings.TrimPrefix(fieldName, "uris/"))
			uris, ok := item.Data["uris"].([]any)
			if err != nil || index < 0 || !ok || index >= len(uris) {
				return "", ErrNotFound
			}
			uri, ok := uris[index].(map[string]any)
			if !ok {
				return "", ErrNotFound
			}
			value, ok := uri["uri"].(string)
			if !ok {
				return "", ErrNotFound
			}
			return value, nil
		}
		// Do not expose unrecognized encrypted/metadata fields as if they were plaintext.
		for _, known := range encryptedFields[item.Type] {
			if known == fieldName {
				if value, ok := item.Data[known].(string); ok {
					return value, nil
				}
				return "", ErrNotFound
			}
		}
		return "", ErrNotFound
	}
	return "", ErrNotFound
}

// Read decrypts only the requested item from a previously validated snapshot.
func (s *Session) Read(snapshot *Snapshot, id, path string) (string, error) {
	if !s.active() {
		return "", ErrAuthentication
	}
	if !validID(id) {
		return "", errors.New("invalid item ID")
	}
	if snapshot == nil || snapshot.profile.ID != s.accountID || snapshot.profile.Key != s.protectedKey {
		return "", errors.New("snapshot does not match session")
	}
	raw, ok := snapshot.items[id]
	if !ok {
		return "", ErrNotFound
	}
	// Resolve only the requested item's organization; personal reads do no RSA work.
	var envelope struct {
		OrganizationID *string `json:"organizationId"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return "", errors.New("invalid item")
	}
	profile := snapshot.profile
	profile.Organizations = nil
	if envelope.OrganizationID != nil {
		for _, org := range snapshot.profile.Organizations {
			if org.ID == *envelope.OrganizationID {
				profile.Organizations = []Organization{org}
				break
			}
		}
	}
	keys, err := s.organizationKeys(context.Background(), profile)
	if err != nil {
		return "", err
	}
	defer keys.close()
	item, err := s.decryptOwnedItem(raw, keys)
	if err != nil {
		return "", err
	}
	return (&Vault{Items: []Item{item}}).Read(id, path)
}

func decryptItemKey(encrypted *string, ownerKey []byte) ([]byte, error) {
	if encrypted == nil {
		return bytes.Clone(ownerKey), nil
	}
	key, err := decrypt(*encrypted, ownerKey)
	if err != nil {
		return nil, err
	}
	if len(key) != 64 {
		clear(key)
		return nil, ErrCipher
	}
	return key, nil
}
