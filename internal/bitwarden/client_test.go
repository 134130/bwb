package bitwarden

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testDevice = "30000000-0000-4000-8000-000000000001"

func TestServerBoundaries(t *testing.T) {
	for _, server := range []string{"http://example.com", "https://user:pass@example.com", "https://example.com?a=b", "https://example.com#x", "file:///tmp/vault", "://"} {
		if _, err := NewClient(server, "user@example.invalid", testDevice); err == nil {
			t.Fatal("accepted unsafe URL", server)
		}
	}
	for _, server := range []string{"http://127.0.0.1:8222", "http://[::1]:8222", "https://example.invalid/base"} {
		if _, err := NewClient(server, " USER@example.invalid ", testDevice); err != nil {
			t.Fatal(err)
		}
	}
	c, err := NewClient("https://vault.bitwarden.eu", "user@example.invalid", testDevice)
	if err != nil || c.identity != "https://identity.bitwarden.eu" || c.api != "https://api.bitwarden.eu" {
		t.Fatal("incorrect cloud endpoints")
	}
}

func TestRequestRejectsRedirectBadJSONAndOversize(t *testing.T) {
	requests := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; w.Write([]byte(`{}`)) }))
	defer target.Close()
	for _, kind := range []string{"redirect", "json", "large", "error"} {
		t.Run(kind, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch kind {
				case "redirect":
					http.Redirect(w, r, target.URL, http.StatusFound)
				case "json":
					io.WriteString(w, `{"secret":`)
				case "large":
					io.CopyN(w, strings.NewReader(strings.Repeat("x", maxResponse+1)), maxResponse+1)
				case "error":
					w.WriteHeader(500)
					io.WriteString(w, "private-server-error")
				}
			}))
			defer server.Close()
			c, _ := NewClient(server.URL, "user@example.invalid", testDevice)
			var result any
			err := c.request(context.Background(), "GET", server.URL, nil, false, "token", &result)
			if err == nil || strings.Contains(err.Error(), "private-server-error") {
				t.Fatal("unsafe response accepted", err)
			}
		})
	}
	if requests != 0 {
		t.Fatal("followed credential redirect")
	}
}

func TestLoginRejectsKDFBeforeTokenRequest(t *testing.T) {
	for _, body := range []string{`{"kdf":1,"kdfIterations":3}`, `{"kdfIterations":600000}`, `{"kdf":0,"kdfIterations":2147483647}`} {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; io.WriteString(w, body) }))
		c, _ := NewClient(server.URL, "user@example.invalid", testDevice)
		session, err := c.Login(context.Background(), []byte("password"))
		server.Close()
		if err == nil || session != nil || requests != 1 {
			t.Fatal("invalid KDF reached token request")
		}
	}
}

func TestRefreshRetryBoundedAndClose(t *testing.T) {
	syncCalls, refreshCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identity/connect/token" {
			refreshCalls++
			io.WriteString(w, `{"access_token":"new","refresh_token":"rotated"}`)
			return
		}
		syncCalls++
		w.WriteHeader(401)
	}))
	defer server.Close()
	c, _ := NewClient(server.URL, "user@example.invalid", testDevice)
	key := make([]byte, 64)
	for i := range key {
		key[i] = 1
	}
	s := &Session{client: c, key: key, access: "expired", refresh: "old"}
	if _, err := s.Sync(context.Background()); !errors.Is(err, ErrAuthentication) {
		t.Fatal(err)
	}
	if syncCalls != 2 || refreshCalls != 1 || s.refresh != "rotated" {
		t.Fatal("incorrect refresh/retry")
	}
	s.Close()
	for _, b := range key {
		if b != 0 {
			t.Fatal("key not cleared")
		}
	}
	if _, err := s.Sync(context.Background()); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed session reused")
	}
	if err := s.Refresh(context.Background()); !errors.Is(err, ErrAuthentication) {
		t.Fatal("closed session refreshed")
	}
}

func TestCanceledRequest(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1", "user@example.invalid", testDevice)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Login(ctx, []byte("password")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSyncRejectsIdentityKeyChangeAndIncompleteResponse(t *testing.T) {
	for _, body := range []string{
		`{"profile":{"id":"20000000-0000-4000-8000-000000000001","email":"other@example.invalid","key":"wrapped"},"folders":[],"ciphers":[]}`,
		`{"profile":{"id":"20000000-0000-4000-8000-000000000001","email":"user@example.invalid","key":"new-key"},"folders":[],"ciphers":[]}`,
		`{"profile":{"id":"20000000-0000-4000-8000-000000000001","email":"user@example.invalid","key":"wrapped"}}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) }))
		c, _ := NewClient(server.URL, "user@example.invalid", testDevice)
		session := &Session{client: c, key: make([]byte, 64), access: "token", protectedKey: "wrapped"}
		vault, err := session.Sync(context.Background())
		session.Close()
		server.Close()
		if err == nil || vault != nil {
			t.Fatal("unsafe or incomplete sync accepted")
		}
	}
}
