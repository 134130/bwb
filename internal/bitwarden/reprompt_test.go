package bitwarden

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestRepromptVerificationAndPermissionBoundary(t *testing.T) {
	_, email, password, token, response := apiKeyFixture(t)
	session, snapshot, err := Unlock(t.Context(), &Cache{Version: 1, Server: "https://example.invalid", Email: email, DeviceID: testDevice, Iterations: token.Iterations, Snapshot: response}, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	const id = "20000000-0000-4000-8000-000000000001"
	if err = session.VerifyPassword(t.Context(), []byte("wrong")); err == nil {
		t.Fatal("wrong password accepted")
	}
	if err = session.VerifyPassword(t.Context(), []byte(password)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err = session.VerifyPassword(ctx, []byte(password)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	authorized, err := snapshot.AllowReprompt(id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Read(authorized, id, "login/password"); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Read(snapshot, id, "login/password"); !errors.Is(err, ErrReprompt) {
		t.Fatal("cached policy mutated", err)
	}
	// The same authorization must never strip a server-provided field restriction.
	var raw map[string]json.RawMessage
	if err = json.Unmarshal(snapshot.items[id], &raw); err != nil {
		t.Fatal(err)
	}
	raw["viewPassword"] = json.RawMessage("false")
	snapshot.items[id], err = json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	authorized, err = snapshot.AllowReprompt(id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Describe(t.Context(), authorized, id); !errors.Is(err, ErrPermission) {
		t.Fatal("permission bypass", err)
	}
	if _, err = session.Read(authorized, id, "login/password"); !errors.Is(err, ErrPermission) {
		t.Fatal("raw permission bypass", err)
	}
}
