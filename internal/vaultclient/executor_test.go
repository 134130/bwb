package vaultclient

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

func TestExecuteRetriesOnceAndReturnsOnlyFinalResult(t *testing.T) {
	calls, auth := 0, 0
	original := errors.New("authentication needed")
	result, err := execute(context.Background(), func() (string, error) {
		calls++
		if calls == 1 {
			return "partial", original
		}
		return "final", nil
	}, func(err error) error {
		auth++
		if err != original {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil || result != "final" || calls != 2 || auth != 1 {
		t.Fatal(result, err, calls, auth)
	}
	calls, auth = 0, 0
	_, err = execute(context.Background(), func() (string, error) { calls++; return "", original }, func(error) error { auth++; return nil })
	if !errors.Is(err, original) || calls != 2 || auth != 1 {
		t.Fatal("unbounded retry", err, calls, auth)
	}
}

func TestExecutorDoesNotAuthenticateOrdinaryOrNoninteractiveErrors(t *testing.T) {
	e := Executor{Prompt: func(context.Context, AuthPrompt) ([]byte, error) { t.Fatal("unexpected prompt"); return nil, nil }}
	for _, err := range []error{
		errors.New("network error"),
		&kit.RPCError{Code: -32000, Message: "login required"},
		&kit.RPCError{Code: vaultapi.StateChangedCode, Message: "vault locked"},
		&kit.RPCError{Code: vaultapi.LoginRequiredCode, Data: []byte(`{"profile":"other"}`)},
		&kit.RPCError{Code: vaultapi.UnlockRequiredCode, Data: []byte(`bad`)},
	} {
		if got := e.authenticate(context.Background(), "default", err); got != err {
			t.Fatal(got)
		}
	}
	e.Prompt = nil
	err := &kit.RPCError{Code: vaultapi.UnlockRequiredCode, Data: []byte(`{"profile":"default"}`)}
	if e.authenticate(context.Background(), "default", err) != err {
		t.Fatal("noninteractive error changed")
	}
}

func TestExecuteCancellationPreventsRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		_, err := execute(ctx, func() (string, error) { calls++; return "", errors.New("locked") }, func(error) error { cancel(); return nil })
		if !errors.Is(err, context.Canceled) || calls != 1 {
			t.Fatal(err, calls)
		}
	})
}
