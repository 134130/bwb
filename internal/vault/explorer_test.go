package vault

import (
	"errors"
	"slices"
	"testing"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/config"
)

func TestProfilesIncludeConfiguredAndPersistedAccounts(t *testing.T) {
	s, _ := persistentService(t)
	defer s.Shutdown(t.Context())
	policy := config.Profile{LockTimeout: "0s"}
	if err := s.ApplyConfig(t.Context(), &config.File{Profiles: map[string]config.Profile{"configured": policy}}); err != nil {
		t.Fatal(err)
	}
	saved, err := s.load(t.Context(), "legacy")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.saveLogin(t.Context(), saved, &bitwarden.Cache{Version: 1, Server: "https://example.invalid", Email: "a@example.invalid"}, credential{Token: []byte("token")}); err != nil {
		t.Fatal(err)
	}
	profiles, err := s.Profiles(t.Context(), Empty{})
	if err != nil || !slices.Equal(profiles.Names, []string{"configured", "legacy"}) {
		t.Fatal(profiles, err)
	}
	_, err = s.List(t.Context(), ListRequest{Profile: "configured", Kind: "item", Background: true})
	var auth *AuthRequired
	if !errors.As(err, &auth) || auth.Action != "login" {
		t.Fatal(err)
	}
	reservation, err := s.BeginAuth(t.Context(), AuthRequest{Profile: auth.Profile, Action: auth.Action, Epoch: auth.Epoch, Version: auth.Version})
	if err != nil || reservation.ID == "" || reservation.Email != "" || reservation.Server != "" {
		t.Fatal(reservation, err)
	}
	if _, err := s.EndAuth(t.Context(), AuthEndRequest{Profile: "configured", ID: reservation.ID}); err != nil {
		t.Fatal(err)
	}
}
