//go:build integration

package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/awnumar/memguard"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestSnapshotIntegration(t *testing.T) {
	server := os.Getenv("BWB_INTEGRATION_URL")
	if server == "" {
		t.Fatal("run through dev/integration/run.py; a disposable server is required")
	}
	var manifest struct {
		Email, Password string
		AccountID       string `json:"account_id"`
		Items, Folders  map[string]string
		Organizations   organizationManifest
	}
	loadJSON(t, "../../testdata/integration/manifest.json", &manifest)
	var fixture struct {
		Items   []map[string]any
		Folders []Folder
	}
	loadJSON(t, "../../testdata/bitwarden.json", &fixture)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	c, err := NewClient(server, manifest.Email, testDevice)
	if err != nil {
		t.Fatal(err)
	}
	if s, err := c.Login(ctx, []byte("wrong password")); !errors.Is(err, ErrAuthentication) || s != nil {
		t.Fatal("wrong password accepted", err)
	}
	password := []byte(manifest.Password)
	session, err := c.Login(ctx, password)
	clear(password)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	// Exercise automatic refresh on a real unauthorized API response as well.
	session.access.Destroy()
	session.access = memguard.NewBufferFromBytes([]byte("invalid-test-access-token"))
	vault, err := session.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	itemCount := len(fixture.Items)
	for _, org := range manifest.Organizations.Entries {
		itemCount += len(org.Items)
	}
	if vault.Profile.ID != manifest.AccountID || len(vault.Items) != itemCount {
		t.Fatal("snapshot identity/count mismatch")
	}
	gotFolders := map[string]string{}
	for _, folder := range vault.Folders {
		gotFolders[folder.ID] = folder.Name
	}
	wantFolders := map[string]string{}
	for _, folder := range fixture.Folders {
		wantFolders[manifest.Folders[folder.ID]] = folder.Name
	}
	if !reflect.DeepEqual(gotFolders, wantFolders) {
		t.Fatal("folder mismatch")
	}
	got := map[string]Item{}
	for _, item := range vault.Items {
		got[item.ID] = item
	}
	types := map[int]bool{}
	for _, expected := range fixture.Items {
		fixtureID := expected["id"].(string)
		item, ok := got[manifest.Items[fixtureID]]
		if !ok {
			t.Fatalf("missing fixture %s", fixtureID)
		}
		types[item.Type] = true
		bytes, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(bytes, &decoded); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"type", "organizationId", "name", "notes", "favorite", "reprompt", "fields"} {
			assertSubset(t, expected[field], decoded[field], fixtureID+"/"+field)
		}
		var folderID any
		if expected["folderId"] != nil {
			folderID = manifest.Folders[expected["folderId"].(string)]
		}
		assertSubset(t, folderID, decoded["folderId"], fixtureID+"/folderId")
		assertSubset(t, expected[payloadNames[item.Type]], decoded["data"], fixtureID+"/payload")
	}
	if len(types) != 8 {
		t.Fatal("missing cipher type")
	}
	snapshot, err := session.SyncSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(session.Cache(snapshot))
	if err != nil {
		t.Fatal(err)
	}
	refresh := session.RefreshToken()
	if bytes.Contains(encoded, []byte(manifest.Password)) || bytes.Contains(encoded, refresh) || bytes.Contains(encoded, []byte("DEMO-only-password-1!")) {
		t.Fatal("cache contains plaintext credentials")
	}
	clear(refresh)
	var cached Cache
	if err := json.Unmarshal(encoded, &cached); err != nil {
		t.Fatal(err)
	}
	unlocked, cachedSnapshot, err := Unlock(ctx, &cached, []byte(manifest.Password))
	if err != nil {
		t.Fatal(err)
	}
	if unlocked.access != nil || unlocked.refresh != nil {
		t.Fatal("offline unlock restored network credentials")
	}
	value, err := unlocked.Read(cachedSnapshot, manifest.Items["20000000-0000-4000-8000-000000000001"], "login/password")
	if err != nil || value != "DEMO-only-password-1!" {
		t.Fatal("cached read mismatch", err)
	}
	unlocked.Close()
	for _, mutate := range []func(*Cache){
		func(c *Cache) { c.Version++ },
		func(c *Cache) { c.KDF = 1 },
		func(c *Cache) { c.Iterations = 2147483647 },
		func(c *Cache) { c.Email = "other@example.invalid" },
		func(c *Cache) { c.Snapshot.Profile.Email = "other@example.invalid" },
		func(c *Cache) { c.Snapshot.Ciphers[0] = json.RawMessage(`{"id":"bad"}`) },
	} {
		var broken Cache
		if err := json.Unmarshal(encoded, &broken); err != nil {
			t.Fatal(err)
		}
		mutate(&broken)
		if leaked, _, err := Unlock(ctx, &broken, []byte(manifest.Password)); err == nil {
			leaked.Close()
			t.Fatal("invalid persisted cache unlocked")
		}
	}
	// A real server ciphertext must not yield any plaintext after tampering.
	var encrypted syncResponse
	if err := c.request(ctx, "GET", c.api+"/sync", nil, false, string(session.access.Bytes()), &encrypted); err != nil {
		t.Fatal(err)
	}
	var entry map[string]any
	if err := json.Unmarshal(encrypted.Ciphers[0], &entry); err != nil {
		t.Fatal(err)
	}
	entry["name"] = "2.invalid|cipher|mac"
	tampered, _ := json.Marshal(entry)
	if _, err := decryptItem(tampered, session.key.Bytes()); err == nil {
		t.Fatal("tampered server ciphertext accepted")
	}
	session.Close()
	if _, err := session.Sync(ctx); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed session usable")
	}
	t.Logf("Go API verified %d items / %d cipher types", len(vault.Items), len(types))
}

func loadJSON(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatal(err)
	}
}

func assertSubset(t *testing.T, want, got any, path string) {
	t.Helper()
	switch expected := want.(type) {
	case map[string]any:
		actual, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("%s: expected object", path)
		}
		for name, value := range expected {
			actualValue, ok := actual[name]
			if !ok {
				t.Fatalf("%s/%s: missing field", path, name)
			}
			assertSubset(t, value, actualValue, path+"/"+name)
		}
	case []any:
		actual, ok := got.([]any)
		if !ok || len(actual) != len(expected) {
			t.Fatalf("%s: array mismatch", path)
		}
		for i, value := range expected {
			assertSubset(t, value, actual[i], path+"/element")
		}
	default:
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("%s: value mismatch", path)
		}
	}
}
