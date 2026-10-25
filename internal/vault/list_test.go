package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/134130/bwb/internal/bitwarden"
)

func TestListFiltersSortAndCursor(t *testing.T) {
	org := &NamedID{ID: "10000000-0000-4000-8000-000000000001", Name: "Work"}
	folder := &NamedID{ID: "10000000-0000-4000-8000-000000000002", Name: "Projects"}
	collection := NamedID{ID: "10000000-0000-4000-8000-000000000003", Name: "Shared"}
	catalog := bitwarden.Catalog{
		"organization": {{ID: org.ID, Name: org.Name}}, "folder": {{ID: folder.ID, Name: folder.Name}},
		"collection": {{ID: collection.ID, Name: collection.Name, Organization: org}},
		"item": {
			{ID: "c", Name: "ALPHA", Type: "login", Organization: org, Collections: []NamedID{collection}, Folder: folder},
			{ID: "b", Name: "alpha", Type: "card", Organization: org, Collections: []NamedID{collection}},
			{ID: "a", Name: "Zulu", Type: "login"},
		},
	}
	req := ListRequest{Profile: "default", Kind: "item", Sort: []string{"name"}}
	page, err := listPage(catalog, "generation", req)
	if err != nil || page.Entries[0].ID != "b" || page.Entries[1].ID != "c" {
		t.Fatal("stable name sort failed", err)
	}
	req.Reverse = true
	page, err = listPage(catalog, "generation", req)
	if err != nil || page.Entries[0].ID != "a" || page.Entries[1].ID != "b" {
		t.Fatal("reverse or tie sort failed", err)
	}
	req.Search, req.OrganizationID, req.CollectionID, req.FolderID = "AlPh", org.ID, collection.ID, folder.ID
	page, err = listPage(catalog, "generation", req)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].ID != "c" {
		t.Fatal("AND filters failed", err)
	}
	req.FolderID = "unknown"
	if _, err := listPage(catalog, "generation", req); err == nil {
		t.Fatal("unknown filter accepted")
	}
	req = ListRequest{Kind: "item", Sort: []string{"type", "name"}}
	page, err = listPage(catalog, "generation", req)
	if err != nil || page.Entries[0].ID != "b" || page.Entries[1].ID != "c" {
		t.Fatal("multi-key sort failed", err)
	}
	for _, mutate := range []func(*ListRequest){
		func(r *ListRequest) { r.Kind = "unknown" }, func(r *ListRequest) { r.CollectionID = "invalid" },
		func(r *ListRequest) { r.Sort = []string{"name", "name"} }, func(r *ListRequest) { r.Sort = []string{"collection"} },
		func(r *ListRequest) { r.Kind = "folder"; r.OrganizationID = org.ID },
	} {
		bad := ListRequest{Kind: "item"}
		mutate(&bad)
		if err := ValidateList(bad); err == nil {
			t.Fatal("invalid query accepted", bad)
		}
	}
}

func TestListByteBoundedPagination(t *testing.T) {
	catalog := bitwarden.Catalog{"item": {}}
	for i := range 300 {
		catalog["item"] = append(catalog["item"], ListEntry{ID: fmt.Sprintf("%04d", i), Name: strings.Repeat("\x01", 100)})
	}
	req := ListRequest{Profile: "default", Kind: "item"}
	ids := map[string]bool{}
	pages := 0
	firstCursor := ""
	for {
		page, err := listPage(catalog, "generation", req)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(page)
		if len(encoded) > 48<<10 || len(page.Entries) == 0 {
			t.Fatal("page exceeded budget or stalled")
		}
		for _, e := range page.Entries {
			if ids[e.ID] {
				t.Fatal("duplicate item across pages")
			}
			ids[e.ID] = true
		}
		pages++
		if firstCursor == "" {
			firstCursor = page.NextCursor
		}
		if page.NextCursor == "" {
			break
		}
		req.Cursor = page.NextCursor
	}
	if len(ids) != 300 || pages < 2 {
		t.Fatal("pagination lost entries")
	}
	req.Cursor = firstCursor
	if _, err := listPage(catalog, "other-generation", req); !errors.Is(err, ErrStateChanged) {
		t.Fatal("stale cursor accepted", err)
	}
	req.Search = "changed"
	if _, err := listPage(catalog, "generation", req); !errors.Is(err, ErrStateChanged) {
		t.Fatal("cursor reused for other query", err)
	}
	req = ListRequest{Kind: "item"}
	catalog["item"] = []ListEntry{{ID: "one", Name: strings.Repeat("\x00", 10000)}}
	if _, err := listPage(catalog, "generation", req); err == nil {
		t.Fatal("oversized entry silently truncated")
	}
	catalog["item"] = nil
	page, err := listPage(catalog, "generation", req)
	if err != nil || page.Entries == nil || len(page.Entries) != 0 || page.NextCursor != "" {
		t.Fatal("empty result contract", err)
	}
}

func TestLockDropsCatalogAndCancelsQueuedList(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := &Service{}
		ctx, p, done, err := s.begin(context.Background(), "default", true)
		if err != nil {
			t.Fatal(err)
		}
		p.catalog = bitwarden.Catalog{"item": {{ID: "secret-name"}}}
		p.catalogID = "old"
		listed := make(chan error, 1)
		go func() {
			_, err := s.List(context.Background(), ListRequest{Profile: "default", Kind: "item"})
			listed <- err
		}()
		synctest.Wait()
		locked := make(chan error, 1)
		go func() { _, err := s.Lock(context.Background(), ProfileRequest{Profile: "default"}); locked <- err }()
		synctest.Wait()
		if ctx.Err() == nil {
			t.Fatal("lock did not cancel current operation")
		}
		done()
		synctest.Wait()
		if err := <-listed; !errors.Is(err, ErrStateChanged) {
			t.Fatal("queued list survived lock", err)
		}
		if err := <-locked; err != nil {
			t.Fatal(err)
		}
		if p.catalog != nil || p.catalogID != "" {
			t.Fatal("lock retained discovery index")
		}
	})
}
