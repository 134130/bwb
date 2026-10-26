package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestItemNameResolutionAndOfflineDetails(t *testing.T) {
	s, _, request, _, requests := apiService(t)
	ctx := t.Context()
	if _, err := s.LoginAPIKey(ctx, request()); err != nil {
		t.Fatal(err)
	}
	before := requests.Load()
	result, err := s.Item(ctx, ItemRequest{Profile: "default", Item: "normal ITEM"})
	if err != nil || result.Item.Name != "Normal item" || len(result.Fields) == 0 {
		t.Fatal(result, err)
	}
	value, err := s.Item(ctx, ItemRequest{Profile: "default", Item: result.Item.ID, Field: "password"})
	if err != nil || *value.Fields[0].Value != "PUBLIC-TEST-PASSWORD" {
		t.Fatal(err)
	}
	if _, err := s.Item(ctx, ItemRequest{Profile: "default", Item: "Normal"}); err == nil {
		t.Fatal("partial name silently selected")
	}
	if requests.Load() != before {
		t.Fatal("item details contacted server")
	}
	if _, err := s.Lock(ctx, ProfileRequest{Profile: "default"}); err != nil {
		t.Fatal(err)
	}
	_, err = s.Item(ctx, ItemRequest{Profile: "default", Item: "Normal item"})
	var auth *AuthRequired
	if !errors.As(err, &auth) || auth.Action != "unlock" {
		t.Fatal("item cannot recover authentication", err)
	}
}

func TestResolveItemAmbiguity(t *testing.T) {
	entries := []ListEntry{{ID: "20000000-0000-4000-8000-000000000001", Name: "Same", Organization: &NamedID{Name: "One"}}, {ID: "20000000-0000-4000-8000-000000000002", Name: "Same", Organization: &NamedID{Name: "Two"}}}
	_, err := resolveItem(entries, "same")
	if err == nil || !strings.Contains(err.Error(), entries[0].ID) || !strings.Contains(err.Error(), "One") || !strings.Contains(err.Error(), "Two") {
		t.Fatal("ambiguous name missing candidates", err)
	}
	if entry, err := resolveItem(entries, entries[1].ID); err != nil || entry.ID != entries[1].ID {
		t.Fatal("ID did not disambiguate", err)
	}
}
