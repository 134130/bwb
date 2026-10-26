package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/google/uuid"
)

type ItemRequest struct {
	Profile string `json:"profile"`
	Item    string `json:"item"`
	Field   string `json:"field,omitempty"`
	Reveal  bool   `json:"reveal,omitempty"`
	TOTP    bool   `json:"totp,omitempty"`
}

type ItemResponse struct {
	Item ListEntry `json:"item"`
	bitwarden.ItemDetails
}

func (s *Service) loadCatalog(ctx context.Context, p *profile) error {
	if p.catalog != nil {
		return nil
	}
	catalog, err := p.session.Catalog(ctx, p.snapshot)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p.catalog, p.catalogID = catalog, uuid.NewString()
	return nil
}

func resolveItem(entries []ListEntry, selector string) (ListEntry, error) {
	id, parseErr := uuid.Parse(selector)
	byID := parseErr == nil && id != uuid.Nil && id.String() == selector
	matches := []ListEntry{}
	for _, entry := range entries {
		if byID && entry.ID == selector || !byID && strings.EqualFold(entry.Name, selector) {
			matches = append(matches, entry)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return ListEntry{}, fmt.Errorf("item %q not found; use item list --search to find an ID", selector)
	}
	var candidates strings.Builder
	for _, entry := range matches[:min(8, len(matches))] {
		fmt.Fprintf(&candidates, "\n  %s", entry.ID)
		if entry.Organization != nil {
			fmt.Fprintf(&candidates, "  organization=%q", entry.Organization.Name)
		}
		for _, collection := range entry.Collections {
			fmt.Fprintf(&candidates, "  collection=%q", collection.Name)
		}
	}
	return ListEntry{}, fmt.Errorf("%d items match %q; use an ID (up to 8 candidates):%s", len(matches), selector, candidates.String())
}

// Item resolves and reads under one profile operation so metadata and values share a snapshot.
func (s *Service) Item(ctx context.Context, req ItemRequest) (ItemResponse, error) {
	if req.Item == "" || len(req.Item) > 4096 || len(req.Field) > 4096 || req.TOTP && (req.Field != "" || req.Reveal) || req.Reveal && req.Field != "" {
		return ItemResponse{}, errors.New("invalid item request")
	}
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return ItemResponse{}, err
	}
	defer done()
	if p.session == nil {
		return ItemResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	if err := s.loadCatalog(ctx, p); err != nil {
		return ItemResponse{}, err
	}
	entry, err := resolveItem(p.catalog["item"], req.Item)
	if err != nil {
		return ItemResponse{}, err
	}
	result := ItemResponse{Item: entry}
	if req.TOTP {
		seed, err := p.session.Read(p.snapshot, entry.ID, "login/totp")
		if err != nil {
			return ItemResponse{}, fmt.Errorf("read TOTP seed: %w", err)
		}
		code, err := bitwarden.TOTP(seed, time.Now())
		if err != nil {
			return ItemResponse{}, err
		}
		result.Fields = []bitwarden.ItemField{{Key: "totp", Label: "Authentication code", Sensitive: true, State: "visible", Value: &code}}
	} else {
		result.ItemDetails, err = p.session.Inspect(ctx, p.snapshot, entry.ID, req.Field, req.Reveal)
		if err != nil {
			return ItemResponse{}, err
		}
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return ItemResponse{}, err
	}
	if len(encoded) > 48000 {
		return ItemResponse{}, errors.New("item exceeds RPC output limit; select one field with --field")
	}
	if err := ctx.Err(); err != nil {
		return ItemResponse{}, err
	}
	if err := s.touch(ctx, p); err != nil {
		return ItemResponse{}, err
	}
	return result, nil
}
