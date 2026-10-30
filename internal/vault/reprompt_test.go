package vault

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
)

func TestRepromptScopedReadAndPolicy(t *testing.T) {
	s, _, request, _, requests := apiService(t)
	login := request()
	password := append([]byte(nil), login.Password...)
	if _, err := s.LoginAPIKey(t.Context(), login); err != nil {
		t.Fatal(err)
	}
	const id = "20000000-0000-4000-8000-000000000001"
	before := requests.Load()
	req := ItemRequest{Profile: "default", Item: id, Field: "password"}
	if _, err := s.Item(t.Context(), req); !errors.Is(err, bitwarden.ErrReprompt) {
		t.Fatal(err)
	}
	if _, err := s.Reprompt(t.Context(), RepromptRequest{Profile: "default", Item: id, Password: []byte("wrong")}); err == nil {
		t.Fatal("wrong password accepted")
	}
	grant, err := s.Reprompt(t.Context(), RepromptRequest{Profile: "default", Item: id, Password: append([]byte(nil), password...)})
	if err != nil {
		t.Fatal(err)
	}
	req.RepromptToken = grant.Token
	if _, err = s.Item(t.Context(), req); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Read(t.Context(), ReadRequest{Profile: "default", ID: id, Field: "login/password", RepromptToken: grant.Token}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Describe(t.Context(), DescribeRequest{Profile: "default", ID: id, RepromptToken: grant.Token}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Item(t.Context(), ItemRequest{Profile: "default", Item: id, Reveal: true}); !errors.Is(err, bitwarden.ErrReprompt) {
		t.Fatal("grant shared implicitly", err)
	}
	if requests.Load() != before {
		t.Fatal("reprompt contacted server")
	}
	policy := &config.File{Version: 1, Profiles: map[string]config.Profile{"default": {InsecureSkipReprompt: true}}}
	if err = s.ApplyConfig(t.Context(), policy); err != nil {
		t.Fatal(err)
	}
	req.RepromptToken = ""
	if _, err = s.Item(t.Context(), req); err != nil {
		t.Fatal("skip not applied", err)
	}
	// Skip removes only reprompt; the server's field permission must survive.
	current := s.profiles["default"]
	original := current.snapshot
	cache := current.session.Cache(original)
	cache.Snapshot.Ciphers = append([]json.RawMessage(nil), cache.Snapshot.Ciphers...)
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(cache.Snapshot.Ciphers[0], &envelope); err != nil {
		t.Fatal(err)
	}
	envelope["viewPassword"] = json.RawMessage("false")
	cache.Snapshot.Ciphers[0], err = json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	restored, restricted, err := bitwarden.Unlock(t.Context(), cache, password)
	if err != nil {
		t.Fatal(err)
	}
	restored.Close()
	current.snapshot = restricted
	if _, err = s.Item(t.Context(), req); !errors.Is(err, bitwarden.ErrPermission) {
		t.Fatal("skip bypassed field restriction", err)
	}
	if _, err = s.Reprompt(t.Context(), RepromptRequest{Profile: "default", Item: id, Password: append([]byte(nil), password...)}); !errors.Is(err, bitwarden.ErrPermission) {
		t.Fatal("password bypassed field restriction", err)
	}
	current.snapshot = original
	p := policy.Profiles["default"]
	p.InsecureSkipReprompt = false
	policy.Profiles["default"] = p
	if err = s.ApplyConfig(t.Context(), policy); err != nil {
		t.Fatal(err)
	}
	req.RepromptToken = grant.Token
	if _, err = s.Item(t.Context(), req); !errors.Is(err, bitwarden.ErrReprompt) {
		t.Fatal("reload kept grant", err)
	}
}

func TestRepromptWatchCancellationAndExpiry(t *testing.T) {
	s, _, request, _, _ := apiService(t)
	login := request()
	password := append([]byte(nil), login.Password...)
	if _, err := s.LoginAPIKey(t.Context(), login); err != nil {
		t.Fatal(err)
	}
	const id = "20000000-0000-4000-8000-000000000001"
	synctest.Test(t, func(t *testing.T) {
		defer func() { _, _ = s.Lock(context.Background(), ProfileRequest{Profile: "default"}) }()
		grant, err := s.Reprompt(t.Context(), RepromptRequest{Profile: "default", Item: id, Password: append([]byte(nil), password...)})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		go func() { _, _ = s.WatchReprompt(ctx, RepromptWatchRequest{Profile: "default", Token: grant.Token}) }()
		synctest.Wait()
		time.Sleep(AuthTimeout + time.Second)
		req := ItemRequest{Profile: "default", Item: id, Field: "password", RepromptToken: grant.Token}
		if _, err = s.Item(t.Context(), req); err != nil {
			t.Fatal("live watch expired", err)
		}
		cancel()
		synctest.Wait()
		if _, err = s.Item(t.Context(), req); !errors.Is(err, bitwarden.ErrReprompt) {
			t.Fatal("disconnected grant accepted", err)
		}
		grant, err = s.Reprompt(t.Context(), RepromptRequest{Profile: "default", Item: id, Password: append([]byte(nil), password...)})
		if err != nil {
			t.Fatal(err)
		}
		req.RepromptToken = grant.Token
		time.Sleep(AuthTimeout)
		if _, err = s.Item(t.Context(), req); !errors.Is(err, bitwarden.ErrReprompt) {
			t.Fatal("abandoned grant accepted", err)
		}
	})
}
