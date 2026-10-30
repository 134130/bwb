package bitwarden

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
)

func TestArgon2Golden(t *testing.T) {
	// Independent Python cryptography Argon2id + HKDF/PBKDF2 output.
	for _, tc := range []struct {
		kdf       kdfConfig
		key, hash string
	}{
		{kdfConfig{1, 3, 64, 4}, "92c6dea0cb40cc0ba6475e27c241b65a80dd950cd0e3abf7a4d117c4268e29424b8bc813a7dfede579bfc7b58392487788fe90dcd0f4c2260b124fca19278c0c", "8mPNDU83hFm2cScU/uOuNtFgjFeq7x1q6XW5MteHcc4="},
		{kdfConfig{1, 2, 16, 1}, "908dd540ac4e6eebaf4447f9e77b546692392b9e248f56c56020a118eeee9d216e64855f01bc6edd22b895b3a93752bbc86ff7dc236d0ee9d5c49c9e284ce60e", "NzfpAY42lPGVYEiesfbQEWUX0K4ciei1bL90s7CuNR8="},
	} {
		key, hash, err := derive(t.Context(), "bwb-argon2id@example.invalid", []byte("PUBLIC-INTEGRATION-ONLY-not-a-real-password!"), tc.kdf)
		if err != nil || hex.EncodeToString(key) != tc.key || hash != tc.hash {
			t.Fatal("Argon2 golden mismatch", err)
		}
		clear(key)
	}
}

// Bitwarden SDK keys/kdf.rs test_master_key_derive_argon2, then Python HKDF/auth hash.
func TestArgon2SDKVector(t *testing.T) {
	key, hash, err := derive(t.Context(), "test_key", []byte("67t9b5g67$%Dh89n"), kdfConfig{1, 4, 32, 2})
	defer clear(key)
	if err != nil || hex.EncodeToString(key) != "ecfda679cf7c62952a8d61e2cf47ad3c0a00b8fffc573e20bca6addf929fdedbd6904cade16a8483ad3886f1dfe3a5a192256fce9b18e09786bdca001b958315" || hash != "r3Eince9qETXVbhwh9B2OSVoJZLxvqHS2ig2kCocNrg=" {
		t.Fatal("SDK vector mismatch", err)
	}
}

func TestArgon2Bounds(t *testing.T) {
	for _, k := range []kdfConfig{
		{2, 3, 64, 4}, {1, 0, 64, 4}, {1, 1, 64, 4}, {1, 11, 64, 4},
		{1, 3, -1, 4}, {1, 3, 0, 4}, {1, 3, 15, 4}, {1, 3, 1025, 4}, {1, 3, 2147483647, 4},
		{1, 3, 64, 0}, {1, 3, 64, 17}, {1, 3, 64, 256},
	} {
		if key, hash, err := derive(t.Context(), "x", []byte("y"), k); err == nil || key != nil || hash != "" {
			t.Fatalf("accepted invalid KDF: %+v", k)
		}
	}
	if err := (kdfConfig{1, 10, 1024, 16}).validate(); err != nil {
		t.Fatal(err)
	}
}

func TestArgon2WaitCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Channel must belong to the bubble for a durably blocked waiter.
		original := argonWork
		argonWork = make(chan struct{}, 1)
		defer func() { argonWork = original }()
		argonWork <- struct{}{}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() { _, _, err := derive(ctx, "x", []byte("y"), kdfConfig{1, 3, 64, 4}); done <- err }()
		synctest.Wait()
		select {
		case err := <-done:
			t.Fatalf("derivation bypassed occupied slot: %v", err)
		default:
		}
		cancel()
		synctest.Wait()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if len(argonWork) != 1 {
			t.Fatal("waiter released someone else's slot")
		}
		<-argonWork
		if key, _, err := derive(ctx, "x", []byte("y"), kdfConfig{1, 3, 64, 4}); !errors.Is(err, context.Canceled) || key != nil {
			t.Fatal("canceled derivation returned key", err)
		}
	})
}

func TestAPIKeyRefreshDetectsAllKDFChanges(t *testing.T) {
	key, email, password, token, _ := apiKeyFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(token) }))
	defer server.Close()
	client, _ := NewClient(server.URL, email, testDevice)
	session, err := client.LoginAPIKey(t.Context(), []byte(password), key)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	original := token
	for _, change := range []func(){
		func() { token.Iterations++ },
		func() { token.Memory = 64 },
		func() { token.Parallelism = 4 },
		func() { kind := 1; token.KDF = &kind; token.Iterations = 3; token.Memory = 64; token.Parallelism = 4 },
	} {
		token = original
		change()
		if err := session.Refresh(t.Context()); !errors.Is(err, ErrKeyChanged) {
			t.Fatal("KDF change not detected", err)
		}
	}
}
