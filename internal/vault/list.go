package vault

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/google/uuid"
)

type ListEntry = bitwarden.ListEntry
type NamedID = bitwarden.NamedID

type ListRequest struct {
	Background     bool     `json:"background,omitempty"`
	Profile        string   `json:"profile"`
	Kind           string   `json:"kind"`
	Search         string   `json:"search,omitempty"`
	Sort           []string `json:"sort,omitempty"`
	Reverse        bool     `json:"reverse,omitempty"`
	OrganizationID string   `json:"organizationId,omitempty"`
	FolderID       string   `json:"folderId,omitempty"`
	CollectionID   string   `json:"collectionId,omitempty"`
	Cursor         string   `json:"cursor,omitempty"`
}

type ListResponse struct {
	Entries    []ListEntry `json:"entries"`
	NextCursor string      `json:"nextCursor,omitempty"`
}

type listCursor struct {
	Generation string `json:"generation"`
	Query      string `json:"query"`
	Offset     int    `json:"offset"`
}

// ValidateList checks both local CLI input and untrusted RPC requests.
func ValidateList(req ListRequest) error {
	switch req.Kind {
	case "item", "folder", "organization", "collection":
	default:
		return errors.New("invalid list kind")
	}
	if len(req.Search) > 4096 || len(req.Cursor) > 2048 {
		return errors.New("list input too long")
	}
	for _, id := range []string{req.OrganizationID, req.FolderID, req.CollectionID} {
		if id == "" {
			continue
		}
		parsed, err := uuid.Parse(id)
		if err != nil || parsed == uuid.Nil || parsed.String() != id {
			return errors.New("invalid filter ID")
		}
	}
	if req.Kind != "item" && (req.FolderID != "" || req.CollectionID != "") {
		return errors.New("folder/collection filters require item list")
	}
	if req.OrganizationID != "" && req.Kind != "item" && req.Kind != "collection" {
		return errors.New("organization filter requires item or collection list")
	}
	seen := map[string]bool{}
	for _, key := range req.Sort {
		if seen[key] {
			return errors.New("duplicate sort key")
		}
		seen[key] = true
		switch key {
		case "name":
		case "type", "folder":
			if req.Kind != "item" {
				return errors.New("type/folder sort requires item list")
			}
		case "organization":
			if req.Kind != "item" && req.Kind != "collection" {
				return errors.New("organization sort requires item or collection list")
			}
		default:
			return errors.New("sort must use name, type, folder, or organization")
		}
	}
	return nil
}

func (p *profile) clearCatalog() { p.catalog = nil; p.catalogID = "" }

func (s *Service) List(ctx context.Context, req ListRequest) (ListResponse, error) {
	if err := ValidateList(req); err != nil {
		return ListResponse{}, err
	}
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return ListResponse{}, err
	}
	defer done()
	if p.session == nil {
		return ListResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	if p.catalog == nil {
		catalog, err := p.session.Catalog(ctx, p.snapshot)
		if err != nil {
			return ListResponse{}, err
		}
		if err := ctx.Err(); err != nil {
			return ListResponse{}, err
		}
		p.catalog, p.catalogID = catalog, uuid.NewString()
	}
	response, err := listPage(p.catalog, p.catalogID, req)
	if err != nil {
		return ListResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return ListResponse{}, err
	}
	if !req.Background {
		if err := s.touch(ctx, p); err != nil {
			return ListResponse{}, err
		}
	}
	return response, nil
}

func listPage(catalog bitwarden.Catalog, generation string, req ListRequest) (ListResponse, error) {
	query := req
	query.Cursor = ""
	encoded, _ := json.Marshal(query)
	sum := sha256.Sum256(encoded)
	queryHash := hex.EncodeToString(sum[:])
	offset := 0
	if req.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(req.Cursor)
		var cursor listCursor
		if err != nil || json.Unmarshal(raw, &cursor) != nil || cursor.Offset < 0 {
			return ListResponse{}, errors.New("invalid list cursor")
		}
		if cursor.Generation != generation || cursor.Query != queryHash {
			return ListResponse{}, ErrStateChanged
		}
		offset = cursor.Offset
	}
	exists := func(kind, id string) bool {
		return slices.ContainsFunc(catalog[kind], func(e ListEntry) bool { return e.ID == id })
	}
	if req.OrganizationID != "" && !exists("organization", req.OrganizationID) {
		return ListResponse{}, errors.New("unknown organization filter")
	}
	if req.FolderID != "" && !exists("folder", req.FolderID) {
		return ListResponse{}, errors.New("unknown folder filter")
	}
	if req.CollectionID != "" {
		i := slices.IndexFunc(catalog["collection"], func(e ListEntry) bool { return e.ID == req.CollectionID })
		if i < 0 {
			return ListResponse{}, errors.New("unknown collection filter")
		}
		collection := catalog["collection"][i]
		if req.OrganizationID != "" && (collection.Organization == nil || collection.Organization.ID != req.OrganizationID) {
			return ListResponse{}, errors.New("collection does not belong to organization filter")
		}
	}
	entries := []ListEntry{}
	search := strings.ToLower(req.Search)
	for _, entry := range catalog[req.Kind] {
		if req.OrganizationID != "" && (entry.Organization == nil || entry.Organization.ID != req.OrganizationID) {
			continue
		}
		if req.FolderID != "" && (entry.Folder == nil || entry.Folder.ID != req.FolderID) {
			continue
		}
		if req.CollectionID != "" && !slices.ContainsFunc(entry.Collections, func(c NamedID) bool { return c.ID == req.CollectionID }) {
			continue
		}
		if !strings.Contains(strings.ToLower(entry.Name), search) {
			continue
		}
		entries = append(entries, entry)
	}
	sortKeys := req.Sort
	if len(sortKeys) == 0 {
		sortKeys = []string{"name"}
	}
	name := func(ref *NamedID) string {
		if ref == nil {
			return ""
		}
		return strings.ToLower(ref.Name)
	}
	slices.SortFunc(entries, func(a, b ListEntry) int {
		for _, key := range sortKeys {
			var order int
			switch key {
			case "name":
				order = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			case "type":
				order = cmp.Compare(a.Type, b.Type)
			case "folder":
				order = cmp.Compare(name(a.Folder), name(b.Folder))
			case "organization":
				order = cmp.Compare(name(a.Organization), name(b.Organization))
			}
			if order != 0 {
				if req.Reverse {
					return -order
				}
				return order
			}
		}
		return cmp.Compare(a.ID, b.ID)
	})
	if offset > len(entries) {
		return ListResponse{}, errors.New("invalid list cursor offset")
	}
	response := ListResponse{Entries: []ListEntry{}}
	for i := offset; i < len(entries); i++ {
		candidate := ListResponse{Entries: append(response.Entries, entries[i])}
		if i+1 < len(entries) {
			raw, _ := json.Marshal(listCursor{Generation: generation, Query: queryHash, Offset: i + 1})
			candidate.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
		}
		raw, err := json.Marshal(candidate)
		if err != nil {
			return ListResponse{}, err
		}
		// Leave space for the transport envelope below daemonkit's 64 KiB limit.
		if len(raw) > 48<<10 {
			if len(response.Entries) == 0 {
				return ListResponse{}, errors.New("list entry exceeds RPC output limit")
			}
			break
		}
		response = candidate
		if len(response.Entries) == 128 {
			break
		}
	}
	return response, nil
}
