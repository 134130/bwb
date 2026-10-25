//go:build !darwin && !linux

package cli

import (
	"context"
	"errors"
)

func terminalPassword(context.Context, authPrompt) ([]byte, error) {
	return nil, errors.New("interactive authentication is supported on macOS and Linux; use --password-stdin with login or unlock")
}
