package bitwarden

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestInspectMasksSecretsAndSelectsDiscoveredKeys(t *testing.T) {
	_, email, password, token, response := apiKeyFixture(t)
	cache := &Cache{Version: 1, Server: "https://example.invalid", Email: email, DeviceID: testDevice, Iterations: token.Iterations, Snapshot: response}
	session, snapshot, err := Unlock(t.Context(), cache, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	const id = "20000000-0000-4000-8000-000000000002"
	detail, err := session.Inspect(t.Context(), snapshot, id, "", false)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(detail)
	if strings.Contains(string(data), "PUBLIC-TEST") || !strings.Contains(string(data), `"value":"demo"`) || !strings.Contains(string(data), `"value":null`) {
		t.Fatal("incorrect masking", string(data))
	}
	for _, field := range detail.Fields {
		selected, err := session.Inspect(t.Context(), snapshot, id, field.Key, false)
		if err != nil || len(selected.Fields) != 1 || selected.Fields[0].Value == nil {
			t.Fatal("advertised field cannot be read", field.Key, err)
		}
	}
	revealed, err := session.Inspect(t.Context(), snapshot, id, "", true)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(revealed)
	if !strings.Contains(string(data), "PUBLIC-TEST-PASSWORD") || !strings.Contains(string(data), "PUBLIC-TEST-NOTE") {
		t.Fatal("reveal did not include secrets")
	}
	if _, err := session.Inspect(t.Context(), snapshot, id, "login/password", false); err == nil || !strings.Contains(err.Error(), "available fields") {
		t.Fatal("legacy path accepted or no discovery hint", err)
	}
	const restrictedID = "20000000-0000-4000-8000-000000000001"
	for _, reveal := range []bool{false, true} {
		result, err := session.Inspect(t.Context(), snapshot, restrictedID, "", reveal)
		if err != nil || result.Restriction == "" || len(result.Fields) != 0 {
			t.Fatal("reprompt restriction lost", err)
		}
	}
	if _, err := session.Inspect(t.Context(), snapshot, restrictedID, "password", false); !errors.Is(err, ErrReprompt) {
		t.Fatal(err)
	}
	if name, err := session.Inspect(t.Context(), snapshot, restrictedID, "name", false); err != nil || *name.Fields[0].Value != "Reprompt item" {
		t.Fatal(err)
	}
	// Duplicate custom names get independent keys and remain individually readable.
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(response.Ciphers[1], &raw)
	var fields []json.RawMessage
	_ = json.Unmarshal(raw["fields"], &fields)
	raw["fields"], _ = json.Marshal(append(fields, fields[0]))
	changed, _ := json.Marshal(raw)
	snapshot.items[id] = changed
	for _, key := range []string{"custom/token#0", "custom/token#1"} {
		result, err := session.Inspect(t.Context(), snapshot, id, key, false)
		if err != nil || *result.Fields[0].Value != "PUBLIC-TEST-FIELD" {
			t.Fatal("duplicate custom field not addressable", key, err)
		}
	}
}
