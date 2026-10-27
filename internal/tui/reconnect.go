package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	kit "github.com/134130/go-daemonkit"
)

// Pending connections remain owned by the program until Update accepts them.
type connected struct {
	client      *kit.Client
	stopCleanup func() bool
}

func (r *Root) reconnect() tea.Cmd {
	connect, lifetime := r.connect, r.ctx
	return r.operation("connected", func(ctx context.Context) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		client, err := connect(ctx)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			client.Close()
			return nil, err
		}
		return &connected{client: client, stopCleanup: context.AfterFunc(lifetime, func() { client.Close() })}, nil
	})
}

// Connection identity survives UI epochs but never follows a replacement client.
type disconnected struct {
	client *kit.Client
	err    error
}

func (r *Root) watchConnection() tea.Cmd {
	client, ctx := r.client, r.ctx
	if client == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return nil
		case <-client.Done():
			return disconnected{client, client.Err()}
		}
	}
}
