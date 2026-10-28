package vaultclient

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

// RepromptSession belongs to one UI item view, or one CLI command.
// Copies of an executor share this capability, never a profile-wide authorization.
type RepromptSession struct {
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	profile, token string
}

func NewRepromptSession(ctx context.Context) *RepromptSession {
	ctx, cancel := context.WithCancel(ctx)
	return &RepromptSession{ctx: ctx, cancel: cancel}
}
func (s *RepromptSession) Close() {
	if s != nil {
		s.cancel()
		s.mu.Lock()
		s.token = ""
		s.mu.Unlock()
	}
}
func (s *RepromptSession) tokenFor(profile string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil || s.profile != profile {
		return ""
	}
	return s.token
}
func RepromptFailure(err error) bool {
	var rpc *kit.RPCError
	return errors.As(err, &rpc) && rpc.Code == vaultapi.RepromptRequiredCode
}

func withReprompt[T any](e Executor, ctx context.Context, profile string, token *string, operation func() (T, error)) (T, error) {
	scope := e.Reprompt
	if scope == nil {
		scope = NewRepromptSession(ctx)
		defer scope.Close()
	}
	*token = scope.tokenFor(profile)
	result, err := execute(ctx, operation, func(err error) error { return e.authenticate(ctx, profile, err) })
	if !RepromptFailure(err) || e.Prompt == nil {
		return result, err
	}
	var rpc *kit.RPCError
	errors.As(err, &rpc)
	var details vaultapi.RepromptDetails
	if json.Unmarshal(rpc.Data, &details) != nil || details.Profile != profile || details.Item == "" {
		return result, errors.New("invalid reprompt challenge")
	}
	promptCtx, cancel := context.WithTimeout(ctx, vaultapi.AuthTimeout)
	defer cancel()
	stop := context.AfterFunc(scope.ctx, cancel)
	defer stop()
	password, err := e.Prompt(promptCtx, AuthPrompt{Action: "reprompt", Profile: profile, Label: "Master Password"})
	defer clear(password)
	if err != nil {
		return result, err
	}
	verified, err := e.Client.Invoke(promptCtx, vaultapi.Reprompt, vaultapi.RepromptRequest{Profile: profile, Item: details.Item, Password: password})
	if err != nil {
		return result, err
	}
	if verified.Token == "" {
		return result, errors.New("missing reprompt capability")
	}
	scope.mu.Lock()
	if scope.ctx.Err() != nil {
		scope.mu.Unlock()
		return result, context.Canceled
	}
	scope.profile, scope.token = profile, verified.Token
	scope.mu.Unlock()
	go func() {
		_, _ = e.Client.Invoke(scope.ctx, vaultapi.WatchReprompt, vaultapi.RepromptWatchRequest{Profile: profile, Token: verified.Token})
		scope.mu.Lock()
		defer scope.mu.Unlock()
		if scope.token == verified.Token {
			scope.token = ""
		}
	}()
	*token = verified.Token
	if err := promptCtx.Err(); err != nil {
		return result, err
	}
	return operation()
}
