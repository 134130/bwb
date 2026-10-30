package vaultclient

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/google/uuid"
)

type AuthPrompt struct {
	Action, Profile, Server, Email, Label string
	Public                                bool
}
type PasswordPrompt func(context.Context, AuthPrompt) ([]byte, error)

type Executor struct {
	Reprompt *RepromptSession
	Client   *kit.Client
	Prompt   PasswordPrompt
}

func (e Executor) Item(ctx context.Context, req vaultapi.ItemRequest) (vaultapi.ItemResponse, error) {
	return withReprompt(e, ctx, req.Profile, &req.RepromptToken, func() (vaultapi.ItemResponse, error) { return e.Client.Invoke(ctx, vaultapi.Item, req) })
}
func (e Executor) Read(ctx context.Context, req vaultapi.ReadRequest) (vaultapi.ReadResponse, error) {
	return withReprompt(e, ctx, req.Profile, &req.RepromptToken, func() (vaultapi.ReadResponse, error) { return e.Client.Invoke(ctx, vaultapi.Read, req) })
}
func (e Executor) Sync(ctx context.Context, req vaultapi.ProfileRequest) (vaultapi.Empty, error) {
	return execute(ctx, func() (vaultapi.Empty, error) { return e.Client.Invoke(ctx, vaultapi.Sync, req) },
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

func (e Executor) authenticate(ctx context.Context, profile string, failure error) error {
	var rpc *kit.RPCError
	if e.Prompt == nil || !errors.As(failure, &rpc) {
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
	if action == "login" && detail.Method != "" {
		return failure
	}
	reservation, err := e.Client.Invoke(ctx, vaultapi.BeginAuth, vaultapi.AuthRequest{Profile: profile, Action: action, Epoch: detail.Epoch, Version: detail.Version})
	if err != nil {
		return err
	}
	if reservation.Ready {
		return nil
	}
	if reservation.ID == "" || action != "login" && (reservation.Email == "" || reservation.Server == "") {
		return errors.New("invalid authentication reservation")
	}
	end := vaultapi.AuthEndRequest{Profile: profile, ID: reservation.ID}
	return RunAuth(ctx, e.Client, end, vaultapi.AuthTimeout, func(authCtx context.Context) error {
		if reservation.Server == "" {
			server, err := e.Prompt(authCtx, AuthPrompt{Action: action, Profile: profile, Email: reservation.Email, Label: "Server URL", Public: true})
			defer clear(server)
			if err != nil {
				return err
			}
			reservation.Server = strings.TrimSpace(string(server))
			if _, err := bitwarden.NewClient(reservation.Server, reservation.Email, uuid.NewString()); err != nil {
				return err
			}
		}
		if reservation.Email == "" {
			email, err := e.Prompt(authCtx, AuthPrompt{Action: action, Profile: profile, Server: reservation.Server, Label: "Email", Public: true})
			defer clear(email)
			if err != nil {
				return err
			}
			reservation.Email = strings.TrimSpace(string(email))
			if reservation.Email == "" {
				return errors.New("email is required for first login")
			}
			if _, err := bitwarden.NewClient(reservation.Server, reservation.Email, uuid.NewString()); err != nil {
				return err
			}
		}
		password, err := e.Prompt(authCtx, AuthPrompt{Action: action, Profile: profile, Server: reservation.Server, Email: reservation.Email})
		defer clear(password)
		if err != nil {
			return err
		}
		if err := authCtx.Err(); err != nil {
			return err
		}
		if action == "unlock" {
			_, err = e.Client.Invoke(authCtx, vaultapi.Unlock, vaultapi.UnlockRequest{Profile: profile, Password: password, AuthID: reservation.ID})
		} else {
			_, err = e.Client.Invoke(authCtx, vaultapi.Login, vaultapi.LoginRequest{Profile: profile, Server: reservation.Server, Email: reservation.Email, Password: password, AuthID: reservation.ID})
		}
		return err
	})
}

// List buffers all pages before output; authentication retries restart the whole query.
func (e Executor) List(ctx context.Context, req vaultapi.ListRequest) ([]vaultapi.ListEntry, error) {
	return execute(ctx, func() ([]vaultapi.ListEntry, error) {
		entries := []vaultapi.ListEntry{}
		pageReq := req
		for {
			response, err := e.Client.Invoke(ctx, vaultapi.List, pageReq)
			if err != nil {
				return nil, err
			}
			entries = append(entries, response.Entries...)
			if response.NextCursor == "" {
				return entries, nil
			}
			if response.NextCursor == pageReq.Cursor {
				return nil, errors.New("list cursor did not advance")
			}
			pageReq.Cursor = response.NextCursor
		}
	}, func(err error) error { return e.authenticate(ctx, req.Profile, err) })
}

func (e Executor) Describe(ctx context.Context, req vaultapi.DescribeRequest) (vaultapi.DescribeResponse, error) {
	return withReprompt(e, ctx, req.Profile, &req.RepromptToken, func() (vaultapi.DescribeResponse, error) { return e.Client.Invoke(ctx, vaultapi.Describe, req) })
}
func (e Executor) TOTP(ctx context.Context, req vaultapi.TOTPRequest) (vaultapi.TOTPResponse, error) {
	return withReprompt(e, ctx, req.Profile, &req.RepromptToken, func() (vaultapi.TOTPResponse, error) { return e.Client.Invoke(ctx, vaultapi.TOTP, req) })
}
