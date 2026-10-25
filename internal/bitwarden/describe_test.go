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

func TestDescribeDecryptsOnlyLabelsAndHonorsRestrictions(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 64)
	session := &Session{key: memguard.NewBufferFromBytes(bytes.Clone(key)), accountID: "account", protectedKey: "protected"}
	defer session.Close()
	label := []byte("service/token")
	n := 16 - len(label)%16
	encrypted := encryptPadded(t, append(label, bytes.Repeat([]byte{byte(n)}, n)...), key)
	const id = "20000000-0000-4000-8000-000000000001"
	item := map[string]any{"id": id, "type": 1, "name": "INVALID-secret-name", "notes": "INVALID-secret-notes", "login": map[string]any{"password": "INVALID-secret-password", "uris": []any{map[string]any{"uri": "INVALID-secret-uri"}}}, "fields": []any{map[string]any{"name": encrypted, "value": "INVALID-secret-value"}}}
	snapshot := &Snapshot{profile: Profile{ID: "account", Key: "protected"}, items: map[string]json.RawMessage{}}
	update := func() { snapshot.items[id], _ = json.Marshal(item) }
	update()
	fields, err := session.Describe(context.Background(), snapshot, id)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(fields)
	if strings.Contains(string(encoded), "INVALID") || !strings.Contains(string(encoded), "fields/service/token") || !strings.Contains(string(encoded), "login/uris/0") {
		t.Fatal(string(encoded))
	}
	item["viewPassword"] = false
	update()
	if _, err := session.Describe(context.Background(), snapshot, id); !errors.Is(err, ErrPermission) {
		t.Fatal(err)
	}
	item["viewPassword"] = true
	item["reprompt"] = 1
	update()
	if _, err := session.Describe(context.Background(), snapshot, id); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}
func TestReadIndexedURI(t *testing.T) {
	const id = "20000000-0000-4000-8000-000000000001"
	vault := Vault{Items: []Item{{ID: id, Type: 1, Data: map[string]any{"uris": []any{map[string]any{"uri": "https://example.invalid"}}}}}}
	value, err := vault.Read(id, "login/uris/0")
	if err != nil || value != "https://example.invalid" {
		t.Fatal(value, err)
	}
	for _, path := range []string{"login/uris/-1", "login/uris/1", "login/uris/x"} {
		if _, err := vault.Read(id, path); err == nil {
			t.Fatal(path)
		}
	}
}
