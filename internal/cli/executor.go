package cli

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

type authPrompt struct{ Action, Profile, Server, Email string }
type passwordPrompt func(context.Context, authPrompt) ([]byte, error)

type executor struct {
	client *kit.Client
	prompt passwordPrompt
}

// Only Read and Sync opt into authentication recovery. Output happens after this returns.
func (e executor) Read(ctx context.Context, req vaultapi.ReadRequest) (vaultapi.ReadResponse, error) {
	return execute(ctx, func() (vaultapi.ReadResponse, error) { return e.client.Invoke(ctx, vaultapi.Read, req) },
		func(err error) error { return e.authenticate(ctx, req.Profile, err) })
}
func (e executor) Sync(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	return execute(ctx, func() (vaultapi.Empty, error) { return e.client.Invoke(ctx, vaultapi.Sync, req) },
		func(err error) error { return e.authenticate(ctx, req.Profile, err) })
}
func execute[T any](ctx context.Context, operation func() (T, error), recoverAuth func(error) error) (T, error) {
	result, err := operation()
	if err == nil {
		return result, nil
	}
	var zero T
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	if err := recoverAuth(err); err != nil {
		return zero, err
	}
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	return operation()
}

func (e executor) authenticate(ctx context.Context, profile string, failure error) error {
	var rpc *kit.RPCError
	if e.prompt == nil || !errors.As(failure, &rpc) {
		return failure
	}
	action := ""
	switch rpc.Code {
	case vaultapi.UnlockRequiredCode:
		action = "unlock"
	case vaultapi.LoginRequiredCode:
		action = "login"
	default:
		return failure
	}
	var detail vaultapi.AuthDetails
	if json.Unmarshal(rpc.Data, &detail) != nil || detail.Profile != profile {
		return failure
	}
	reservation, err := e.client.Invoke(ctx, vaultapi.BeginAuth, vaultapi.AuthRequest{Profile: profile, Action: action, Epoch: detail.Epoch, Version: detail.Version})
	if err != nil {
		return err
	}
	if reservation.Ready {
		return nil
	}
	if reservation.ID == "" || reservation.Server == "" || reservation.Email == "" {
		return errors.New("invalid authentication reservation")
	}
	end := vaultapi.AuthEndRequest{Profile: profile, ID: reservation.ID}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		_, _ = e.client.Invoke(cleanup, vaultapi.EndAuth, end)
	}()
	authCtx, cancel := context.WithTimeout(ctx, vaultapi.AuthTimeout)
	defer cancel()
	promptCtx, cancelPrompt := context.WithCancelCause(authCtx)
	defer cancelPrompt(nil)
	watchCtx, stopWatch := context.WithCancel(authCtx)
	defer stopWatch()
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		result, err := e.client.Invoke(watchCtx, vaultapi.WatchAuth, end)
		if err != nil {
			cancelPrompt(err)
		} else if !result.Succeeded {
			cancelPrompt(errors.New("authentication canceled or vault state changed"))
		}
	}()
	defer func() { stopWatch(); <-watched }()
	password, err := e.prompt(promptCtx, authPrompt{Action: action, Profile: profile, Server: reservation.Server, Email: reservation.Email})
	defer clear(password)
	if cause := context.Cause(promptCtx); cause != nil {
		return cause
	}
	if err != nil {
		return err
	}
	if action == "unlock" {
		_, err = e.client.Invoke(authCtx, vaultapi.Unlock, vaultapi.UnlockRequest{Profile: profile, Password: password, AuthID: reservation.ID})
	} else {
		_, err = e.client.Invoke(authCtx, vaultapi.Login, vaultapi.LoginRequest{Profile: profile, Server: reservation.Server, Email: reservation.Email, Password: password, AuthID: reservation.ID})
	}
	return err
}
