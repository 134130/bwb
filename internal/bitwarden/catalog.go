package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
)

// NamedID keeps names unambiguous without exposing encrypted keys or secret fields.
type NamedID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListEntry struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type,omitempty"`
	Organization *NamedID  `json:"organization,omitempty"`
	Folder       *NamedID  `json:"folder,omitempty"`
	Collections  []NamedID `json:"collections,omitempty"`
	Favorite     bool      `json:"favorite,omitempty"`
}

type Catalog map[string][]ListEntry

// Catalog decrypts only discovery metadata, never passwords, notes or custom fields.
func (s *Session) Catalog(ctx context.Context, snapshot *Snapshot) (Catalog, error) {
	if !s.active() {
		return nil, ErrAuthentication
	}
	if snapshot == nil || snapshot.profile.ID != s.accountID || snapshot.profile.Key != s.protectedKey {
		return nil, errors.New("snapshot does not match session")
	}
	keys, err := s.organizationKeys(ctx, snapshot.profile)
	if err != nil {
		return nil, err
	}
	defer keys.close()
	result := Catalog{"item": {}, "folder": {}, "organization": {}, "collection": {}}
	organizations := map[string]*NamedID{}
	for _, org := range snapshot.profile.Organizations {
		organizations[org.ID] = &NamedID{ID: org.ID, Name: org.Name}
		result["organization"] = append(result["organization"], ListEntry{ID: org.ID, Name: org.Name})
	}
	folders := map[string]*NamedID{}
	for _, folder := range snapshot.response.Folders {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, err := decryptText(folder.Name, s.key.Bytes())
		if err != nil {
			return nil, err
		}
		folders[folder.ID] = &NamedID{ID: folder.ID, Name: name}
		result["folder"] = append(result["folder"], ListEntry{ID: folder.ID, Name: name})
	}
	collections, err := decryptCollections(snapshot.response.Collections, keys)
	if err != nil {
		return nil, err
	}
	collectionNames := map[string]NamedID{}
	for _, collection := range collections {
		collectionNames[collection.ID] = NamedID{ID: collection.ID, Name: collection.Name}
		result["collection"] = append(result["collection"], ListEntry{ID: collection.ID, Name: collection.Name, Organization: organizations[collection.OrganizationID]})
	}
	for _, raw := range snapshot.response.Ciphers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var metadata struct {
			ID, Name                      string
			Type                          int
			OrganizationID, FolderID, Key *string
			CollectionIDs                 []string
			Favorite                      bool
		}
		if json.Unmarshal(raw, &metadata) != nil {
			return nil, errors.New("invalid item metadata")
		}
		entry := ListEntry{ID: metadata.ID, Type: payloadNames[metadata.Type], Favorite: metadata.Favorite, Collections: []NamedID{}}
		ownerKey := s.key
		if metadata.OrganizationID != nil {
			ownerKey = keys[*metadata.OrganizationID]
			entry.Organization = organizations[*metadata.OrganizationID]
			if ownerKey == nil || entry.Organization == nil {
				return nil, errors.New("unknown item organization")
			}
		}
		key, err := decryptItemKey(metadata.Key, ownerKey.Bytes())
		if err != nil {
			return nil, err
		}
		entry.Name, err = decryptText(metadata.Name, key)
		clear(key)
		if err != nil {
			return nil, err
		}
		if metadata.FolderID != nil {
			entry.Folder = folders[*metadata.FolderID]
			if entry.Folder == nil {
				return nil, errors.New("unknown item folder")
			}
		}
		for _, id := range metadata.CollectionIDs {
			name, ok := collectionNames[id]
			if !ok {
				return nil, errors.New("unknown item collection")
			}
			entry.Collections = append(entry.Collections, name)
		}
		result["item"] = append(result["item"], entry)
	}
	return result, nil
}
