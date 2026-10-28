package daemon

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/134130/bwb/internal/vault"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

// rpcHandler translates domain failures only at the transport boundary.
func rpcHandler[Req, Resp any](handler func(context.Context, Req) (Resp, error)) func(context.Context, Req) (Resp, error) {
	return func(ctx context.Context, req Req) (Resp, error) {
		response, err := handler(ctx, req)
		if err == nil {
			return response, nil
		}
		var zero Resp
		// Cancellation has priority over recoverable authentication state.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return zero, err
		}
		if errors.Is(err, vault.ErrStateChanged) {
			return zero, &kit.RPCError{Code: vaultapi.StateChangedCode, Message: err.Error()}
		}
		var reprompt *vault.RepromptRequired
		if errors.As(err, &reprompt) {
			data, _ := json.Marshal(reprompt)
			return zero, &kit.RPCError{Code: vaultapi.RepromptRequiredCode, Message: reprompt.Error(), Data: data}
		}
		var auth *vault.AuthRequired
		if errors.As(err, &auth) {
			code := vaultapi.UnlockRequiredCode
			if auth.Action == "login" {
				code = vaultapi.LoginRequiredCode
			} else if auth.Action != "unlock" {
				return zero, err
			}
			data, _ := json.Marshal(vaultapi.AuthDetails{Profile: auth.Profile, Method: auth.Method, Epoch: auth.Epoch, Version: auth.Version})
			return zero, &kit.RPCError{Code: code, Message: auth.Error(), Data: data}
		}
		return zero, err
	}
}
