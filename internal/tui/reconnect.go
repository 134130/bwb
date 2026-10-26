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
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
