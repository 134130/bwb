package bitwarden

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
)

func TestCatalogDecryptsOnlyMetadata(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 64)
	session := &Session{key: memguard.NewBufferFromBytes(bytes.Clone(key)), accountID: "account", protectedKey: "protected"}
	defer session.Close()
	encrypted := encryptPadded(t, append([]byte("name"), bytes.Repeat([]byte{12}, 12)...), key)
	profile := Profile{ID: "account", Key: "protected"}
	raw, _ := json.Marshal(map[string]any{"id": "20000000-0000-4000-8000-000000000001", "type": 1, "name": encrypted,
		"notes": "INVALID-secret-note", "login": map[string]any{"password": "INVALID-secret-password"},
		"fields": []map[string]any{{"name": "INVALID", "value": "INVALID-secret-field"}}})
	snapshot := &Snapshot{profile: profile, response: syncResponse{Profile: profile, Ciphers: []json.RawMessage{raw}}}
	catalog, err := session.Catalog(context.Background(), snapshot)
	if err != nil || len(catalog["item"]) != 1 || catalog["item"][0].Name != "name" {
		t.Fatal("metadata lookup decrypted secret payload", err)
	}
	encoded, _ := json.Marshal(catalog)
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "password") {
		t.Fatal("catalog retained secret payload")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := session.Catalog(ctx, snapshot); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled catalog succeeded", err)
	}
	snapshot.profile.ID = "other"
	if _, err := session.Catalog(context.Background(), snapshot); err == nil {
		t.Fatal("foreign snapshot accepted")
	}
	session.Close()
	if _, err := session.Catalog(context.Background(), snapshot); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed session listed", err)
	}
}
