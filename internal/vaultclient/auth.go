package vaultclient

import (
	"context"
	"errors"
	"time"

	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

// RunAuth watches the reservation until the client finishes, disconnects, or is canceled.
func RunAuth(ctx context.Context, client *kit.Client, end vaultapi.AuthEndRequest, timeout time.Duration, action func(context.Context) error) error {
	defer func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		_, _ = client.Invoke(cleanup, vaultapi.EndAuth, end)
	}()
	authCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	workCtx, cancelWork := context.WithCancelCause(authCtx)
	defer cancelWork(nil)
	watchCtx, stopWatch := context.WithCancel(authCtx)
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		result, err := client.Invoke(watchCtx, vaultapi.WatchAuth, end)
		if err != nil {
			cancelWork(err)
		} else if !result.Succeeded && !result.Failed {
			cancelWork(errors.New("authentication canceled or vault state changed"))
		}
	}()
	defer func() { stopWatch(); <-watched }()
	err := action(workCtx)
	if cause := context.Cause(workCtx); cause != nil {
		return cause
	}
	return err
}
