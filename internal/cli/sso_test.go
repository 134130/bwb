package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
)

func TestSSOCallbackValidation(t *testing.T) {
	result := make(chan ssoResult, 1)
	handler := ssoCallback("expected-state", result)
	for _, target := range []string{
		"/?code=secret", "/?code=secret&state=wrong", "/favicon.ico?code=secret&state=expected-state",
		"/?state=expected-state", "/?code=&state=expected-state", "/?code=one&code=two&state=expected-state",
		"/?code=secret&state=expected-state&state=wrong", "/?error=denied&code=secret&state=expected-state",
		"/?code=secret&state=expected-state_extra", "/?code=%zz&state=expected-state",
		"/?code=" + strings.Repeat("c", 4097) + "&state=expected-state",
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
		if w.Code != 400 || len(result) != 0 || strings.Contains(w.Body.String(), "secret") {
			t.Fatal("invalid callback accepted or echoed code", target)
		}
	}
	for i, method := range []string{"POST", "GET", "GET"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(method, "/?code=secret&state=expected-state_identifier=work", nil))
		if w.Code != []int{400, 200, 409}[i] {
			t.Fatal("unexpected callback status", w.Code)
		}
		if i == 1 {
			response := <-result
			if string(response.code) != "secret" || response.err != nil || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("invalid accepted callback")
			}
		}
	}
	if len(result) != 0 {
		t.Fatal("replayed callback was accepted")
	}
	denied := make(chan ssoResult, 1)
	w := httptest.NewRecorder()
	ssoCallback("state", denied).ServeHTTP(w, httptest.NewRequest("GET", "/?state=state&error=access_denied&error_description=private-details", nil))
	response := <-denied
	if response.err == nil || len(response.code) != 0 || strings.Contains(w.Body.String()+response.err.Error(), "private-details") {
		t.Fatal("authorization failure leaked details or succeeded")
	}
}

func TestSSOCallbackAcceptsOnlyOneConcurrentResponse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		result := make(chan ssoResult, 1)
		handler := ssoCallback("state", result)
		var accepted atomic.Int32
		for range 8 {
			go func() {
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", "/?code=code&state=state", nil))
				if w.Code == 200 {
					accepted.Add(1)
				}
			}()
		}
		synctest.Wait()
		if accepted.Load() != 1 || len(result) != 1 {
			t.Fatal("concurrent callbacks accepted more than one response")
		}
	})
}

func TestBrowserSSOPKCEAndListenerCleanup(t *testing.T) {
	var challenge, callback string
	code, err := browserSSO(t.Context(), "https://vault.example.invalid/base/", "person+test@example.invalid", "work & test", func(ctx context.Context, uri string) error {
		web, err := url.Parse(uri)
		if err != nil {
			return err
		}
		fragment, _ := url.Parse(web.EscapedFragment())
		query := fragment.Query()
		callback, challenge = query.Get("redirectUri"), query.Get("codeChallenge")
		if web.Host != "vault.example.invalid" || web.Path != "/base/" || fragment.Path != "/sso" ||
			query.Get("clientId") != "cli" || query.Get("identifier") != "work & test" || query.Get("email") != "person+test@example.invalid" ||
			len(query.Get("state")) < 20 || query.Has("codeVerifier") {
			t.Error("invalid browser SSO URL", uri)
		}
		params := url.Values{"state": {query.Get("state") + "_identifier=work & test"}, "code": {"returned-code"}}
		req, _ := http.NewRequestWithContext(ctx, "GET", callback+"/?"+params.Encode(), nil)
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body)
		if response.StatusCode != 200 {
			t.Error("callback failed", response.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(code.Code)
	defer clear(code.Verifier)
	hash := sha256.Sum256(code.Verifier)
	if code.Validate() != nil || string(code.Code) != "returned-code" || code.RedirectURI != callback || base64.RawURLEncoding.EncodeToString(hash[:]) != challenge {
		t.Fatal("code or PKCE verifier not preserved")
	}
	u, _ := url.Parse(callback)
	listener, err := net.Listen("tcp4", "127.0.0.1:"+u.Port())
	if err != nil {
		t.Fatal("callback listener was not closed", err)
	}
	_ = listener.Close()
}

func TestBrowserSSOCancellationAndOpenFailure(t *testing.T) {
	for _, failOpen := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		failure := errors.New("open failed")
		code, err := browserSSO(ctx, "https://example.invalid", "", "", func(context.Context, string) error {
			if failOpen {
				return failure
			}
			cancel()
			return nil
		})
		cancel()
		want := error(context.Canceled)
		if failOpen {
			want = failure
		}
		if !errors.Is(err, want) || len(code.Verifier) != 0 || len(code.Code) != 0 {
			t.Fatal("failed browser login returned credentials", err)
		}
	}
}

func TestSSOFlagsFailBeforeReadingPassword(t *testing.T) {
	isolateConfig(t)
	for _, args := range [][]string{
		{"login", "--sso", "--apikey"}, {"login", "--sso", "--non-interactive"},
		{"login", "--identifier", "work"}, {"login", "--sso", "--identifier", ""},
		{"login", "--sso", "--server", "http://remote.invalid"},
	} {
		root := NewRoot()
		input := strings.NewReader("password\n")
		root.SetIn(input)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		root.SetArgs(append(args, "--password-stdin"))
		if err := root.Execute(); err == nil || input.Len() != len("password\n") {
			t.Fatal("invalid SSO flags reached password input", args, err)
		}
	}
}
