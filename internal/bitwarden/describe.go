package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// FieldInfo contains discoverable labels and read paths, never field values.
type FieldInfo struct {
	Path  string `json:"path"`
	Label string `json:"label"`
}

func (s *Session) Describe(ctx context.Context, snapshot *Snapshot, id string) ([]FieldInfo, error) {
	if !s.active() {
		return nil, ErrAuthentication
	}
	if !validID(id) {
		return nil, errors.New("invalid item ID")
	}
	if snapshot == nil || snapshot.profile.ID != s.accountID || snapshot.profile.Key != s.protectedKey {
		return nil, errors.New("snapshot does not match session")
	}
	raw, ok := snapshot.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	var item Item
	var envelope map[string]json.RawMessage
	if json.Unmarshal(raw, &item) != nil || json.Unmarshal(raw, &envelope) != nil {
		return nil, errors.New("invalid item")
	}
	if item.Reprompt != 0 {
		return nil, ErrUnsupported
	}
	if item.ViewPassword != nil && !*item.ViewPassword {
		return nil, ErrPermission
	}
	if item.OrganizationID != nil && item.ViewPassword == nil {
		return nil, errors.New("missing organization item permissions")
	}
	kind, ok := payloadNames[item.Type]
	if !ok {
		return nil, ErrUnsupported
	}
	var data map[string]json.RawMessage
	if json.Unmarshal(envelope[kind], &data) != nil || data == nil {
		return nil, errors.New("missing item payload")
	}
	fields := []FieldInfo{}
	if item.Notes != nil {
		fields = append(fields, FieldInfo{"notes", "Notes"})
	}
	for _, name := range encryptedFields[item.Type] {
		if value := data[name]; len(value) > 0 && string(value) != "null" {
			fields = append(fields, FieldInfo{kind + "/" + name, name})
		}
	}
	if item.Type == 1 && data["uris"] != nil {
		var uris []map[string]json.RawMessage
		if err := json.Unmarshal(data["uris"], &uris); err != nil {
			return nil, errors.New("invalid login URIs")
		}
		for i, uri := range uris {
			if value := uri["uri"]; len(value) > 0 && string(value) != "null" {
				fields = append(fields, FieldInfo{fmt.Sprintf("login/uris/%d", i), fmt.Sprintf("URI %d", i+1)})
			}
		}
	}
	profile := snapshot.profile
	profile.Organizations = nil
	if item.OrganizationID != nil {
		for _, org := range snapshot.profile.Organizations {
			if org.ID == *item.OrganizationID {
				profile.Organizations = append(profile.Organizations, org)
			}
		}
	}
	keys, err := s.organizationKeys(ctx, profile)
	if err != nil {
		return nil, err
	}
	defer keys.close()
	owner := s.key
	if item.OrganizationID != nil {
		owner = keys[*item.OrganizationID]
		if owner == nil {
			return nil, errors.New("unknown item organization")
		}
	}
	var encryptedKey *string
	if value := envelope["key"]; value != nil {
		if err := json.Unmarshal(value, &encryptedKey); err != nil {
			return nil, err
		}
	}
	key, err := decryptItemKey(encryptedKey, owner.Bytes())
	if err != nil {
		return nil, err
	}
	defer clear(key)
	for _, field := range item.Fields {
		if field.Name == nil {
			continue
		}
		name, err := decryptText(*field.Name, key)
		if err != nil {
			return nil, err
		}
		if field.Value != nil {
			fields = append(fields, FieldInfo{"fields/" + name, name})
		}
	}
	return fields, nil
}
