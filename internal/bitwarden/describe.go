package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// FieldInfo contains discoverable labels and read paths, never field values.
type FieldInfo struct {
	Path      string `json:"path"`
	Key       string `json:"key"`
	Label     string `json:"label"`
	Sensitive bool   `json:"sensitive"`
}

func fieldLabel(name string) string {
	var label strings.Builder
	for i, r := range name {
		if i == 0 {
			r = unicode.ToUpper(r)
		} else if unicode.IsUpper(r) {
			label.WriteByte(' ')
			r = unicode.ToLower(r)
		}
		label.WriteRune(r)
	}
	return label.String()
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
		return nil, ErrReprompt
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
		fields = append(fields, FieldInfo{Path: "notes", Key: "notes", Label: "Notes", Sensitive: true})
	}
	for _, name := range encryptedFields[item.Type] {
		if value := data[name]; len(value) > 0 && string(value) != "null" {
			field := FieldInfo{Path: kind + "/" + name, Key: name, Label: fieldLabel(name), Sensitive: true}
			if kind == "login" && name == "username" || kind == "card" && name == "brand" {
				field.Sensitive = false
			}
			if kind == "login" && name == "totp" {
				field.Key, field.Label = "totp-secret", "TOTP seed"
			}
			fields = append(fields, field)
		}
	}
	if item.Type == 1 && data["uris"] != nil {
		var uris []map[string]json.RawMessage
		if err := json.Unmarshal(data["uris"], &uris); err != nil {
			return nil, errors.New("invalid login URIs")
		}
		for i, uri := range uris {
			if value := uri["uri"]; len(value) > 0 && string(value) != "null" {
				fields = append(fields, FieldInfo{Path: fmt.Sprintf("login/uris/%d", i), Key: fmt.Sprintf("url/%d", i+1), Label: fmt.Sprintf("URL %d", i+1)})
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
	customCounts := map[string]int{}
	for i, field := range item.Fields {
		if field.Name == nil {
			continue
		}
		name, err := decryptText(*field.Name, key)
		if err != nil {
			return nil, err
		}
		if field.Value != nil {
			key := "custom/" + url.PathEscape(name)
			customCounts[key]++
			fields = append(fields, FieldInfo{Path: fmt.Sprintf("custom/%d", i), Key: key, Label: name, Sensitive: field.Type == 1})
		}
	}
	for i := range fields {
		if customCounts[fields[i].Key] > 1 {
			fields[i].Key += "#" + strings.TrimPrefix(fields[i].Path, "custom/")
		}
	}
	return fields, nil
}
