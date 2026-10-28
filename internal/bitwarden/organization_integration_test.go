//go:build integration

package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"slices"
	"testing"
)

type organizationManifest struct {
	RestrictedEmail string `json:"restricted_email"`
	Entries         []struct {
		ID               string
		Collections      []string
		RestrictedMember string `json:"restricted_member"`
		Items            []struct {
			ID            string
			CollectionIDs []string `json:"collection_ids"`
			FolderID      *string  `json:"folder_id"`
		}
	}
}

func TestOrganizationSnapshot(t *testing.T) {
	var manifest struct {
		Email, Password string
		Items           map[string]string
		Organizations   organizationManifest
	}
	loadJSON(t, "../../testdata/integration/manifest.json", &manifest)
	var fixture struct {
		Organizations []struct {
			Name        string
			Collections []string
		}
		Items []struct{ Name, Password string }
	}
	loadJSON(t, "../../testdata/organizations.json", &fixture)
	ctx := context.Background()
	server := os.Getenv("BWB_INTEGRATION_URL")
	ownerClient, err := NewClient(server, manifest.Email, testDevice)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := ownerClient.Login(ctx, []byte(manifest.Password))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(owner.Close)
	restrictedClient, err := NewClient(server, manifest.Organizations.RestrictedEmail, testDevice)
	if err != nil {
		t.Fatal(err)
	}
	restricted, err := restrictedClient.Login(ctx, []byte(manifest.Password))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(restricted.Close)
	for _, session := range []*Session{owner, restricted} {
		limited := session == restricted
		snapshot, err := session.SyncSnapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		vault, err := session.decryptSnapshot(ctx, snapshot.response)
		if err != nil {
			t.Fatal(err)
		}
		collectionCount := 0
		for _, org := range fixture.Organizations {
			collectionCount += len(org.Collections)
		}
		if len(vault.Collections) != collectionCount || len(vault.Profile.Organizations) != len(fixture.Organizations) {
			t.Fatal("missing collections/organizations")
		}
		for n, org := range manifest.Organizations.Entries {
			for k, cid := range org.Collections {
				i := slices.IndexFunc(vault.Collections, func(c Collection) bool { return c.ID == cid })
				if i < 0 || vault.Collections[i].Name != fixture.Organizations[n].Collections[k] || vault.Collections[i].OrganizationID != org.ID {
					t.Fatal("collection mismatch")
				}
			}
			for k, expected := range org.Items {
				i := slices.IndexFunc(vault.Items, func(item Item) bool { return item.ID == expected.ID })
				if i < 0 {
					t.Fatal("missing organization item")
				}
				item := vault.Items[i]
				gotIDs, wantIDs := slices.Clone(item.CollectionIDs), slices.Clone(expected.CollectionIDs)
				slices.Sort(gotIDs)
				slices.Sort(wantIDs)
				if item.OrganizationID == nil || *item.OrganizationID != org.ID || !slices.Equal(gotIDs, wantIDs) || item.Name != fixture.Items[k].Name {
					t.Fatal("item relationship mismatch")
				}
				if !limited && !reflect.DeepEqual(item.FolderID, expected.FolderID) {
					t.Fatal("folder/collection coexistence lost")
				}
				value, err := session.Read(snapshot, item.ID, "login/password")
				if limited && k == 1 {
					if !errors.Is(err, ErrPermission) || value != "" {
						t.Fatal("hidden password exposed", err)
					}
				} else if err != nil || value != fixture.Items[k].Password {
					t.Fatal("organization password mismatch", err)
				}
			}
		}
		encoded, err := json.Marshal(session.Cache(snapshot))
		if err != nil {
			t.Fatal(err)
		}
		var cache Cache
		if err := json.Unmarshal(encoded, &cache); err != nil {
			t.Fatal(err)
		}
		offline, restored, err := Unlock(ctx, &cache, []byte(manifest.Password))
		if err != nil {
			t.Fatal(err)
		}
		for _, org := range manifest.Organizations.Entries {
			value, err := offline.Read(restored, org.Items[0].ID, "login/password")
			if err != nil || value != fixture.Items[0].Password {
				t.Fatal("offline organization read failed", err)
			}
		}
		keys, err := offline.organizationKeys(ctx, restored.profile)
		if err != nil {
			t.Fatal(err)
		}
		keys.close()
		for _, key := range keys {
			if key.IsAlive() {
				t.Fatal("organization key not destroyed")
			}
		}
		offline.Close()
		if value, err := offline.Read(restored, manifest.Organizations.Entries[0].Items[0].ID, "login/password"); err == nil || value != "" {
			t.Fatal("closed organization session usable")
		}
		if !limited {
			// Cache v1 without organization fields remains usable after upgrading.
			var legacy Cache
			if err := json.Unmarshal(encoded, &legacy); err != nil {
				t.Fatal(err)
			}
			legacy.Snapshot.Profile.Organizations = nil
			legacy.Snapshot.Profile.PrivateKey = ""
			legacy.Snapshot.Collections = nil
			legacy.Snapshot.Ciphers = slices.DeleteFunc(legacy.Snapshot.Ciphers, func(raw json.RawMessage) bool {
				var item Item
				if err := json.Unmarshal(raw, &item); err != nil {
					t.Fatal(err)
				}
				return item.OrganizationID != nil
			})
			oldSession, oldSnapshot, err := Unlock(ctx, &legacy, []byte(manifest.Password))
			if err != nil {
				t.Fatal("legacy personal cache rejected", err)
			}
			if len(oldSnapshot.items) != len(manifest.Items) {
				t.Fatal("legacy personal items lost")
			}
			oldSession.Close()
		}
		// Broken key or cross-organization relationship must reject the entire replacement.
		for _, mutate := range []func(*Cache){
			func(c *Cache) { c.Snapshot.Profile.Organizations[0].Key = c.Snapshot.Profile.Organizations[1].Key },
			func(c *Cache) { c.Snapshot.Profile.PrivateKey = "2.invalid" },
			func(c *Cache) { c.Snapshot.Collections[0].OrganizationID = "unknown" },
			func(c *Cache) { c.Snapshot.Collections = nil },
			func(c *Cache) {
				c.Snapshot.Profile.Organizations = append(c.Snapshot.Profile.Organizations, c.Snapshot.Profile.Organizations[0])
			},
		} {
			var broken Cache
			if err := json.Unmarshal(encoded, &broken); err != nil {
				t.Fatal(err)
			}
			mutate(&broken)
			if leaked, _, err := Unlock(ctx, &broken, []byte(manifest.Password)); err == nil {
				leaked.Close()
				t.Fatal("corrupt organization cache accepted")
			}
		}
	}
	// Change a real membership: cached data remains available until the next successful sync.
	org := manifest.Organizations.Entries[0]
	endpoint := ownerClient.api + "/organizations/" + org.ID + "/users/" + org.RestrictedMember
	update := func(restore bool) error {
		collections := []map[string]any{}
		if restore {
			for n, id := range org.Collections {
				collections = append(collections, map[string]any{"id": id, "readOnly": true, "hidePasswords": n == 0, "manage": false})
			}
		}
		payload, _ := json.Marshal(map[string]any{"type": 2, "collections": collections, "groups": []string{}})
		req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(owner.access.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		response, err := ownerClient.http.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body)
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("membership update HTTP %d", response.StatusCode)
		}
		return nil
	}
	before, err := restricted.SyncSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := update(true); err != nil {
			t.Error("restore membership:", err)
		}
	})
	if err := update(false); err != nil {
		t.Fatal(err)
	}
	if _, err := restricted.Read(before, org.Items[0].ID, "login/password"); err != nil {
		t.Fatal("offline snapshot unexpectedly revoked", err)
	}
	after, err := restricted.SyncSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if value, err := restricted.Read(after, org.Items[0].ID, "login/password"); !errors.Is(err, ErrNotFound) || value != "" {
		t.Fatal("revoked item survived sync", err)
	}
	other := manifest.Organizations.Entries[1]
	if value, err := restricted.Read(after, other.Items[0].ID, "login/password"); err != nil || value != fixture.Items[0].Password {
		t.Fatal("other organization was affected", err)
	}
}
