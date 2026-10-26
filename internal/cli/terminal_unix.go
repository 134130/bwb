//go:build darwin || linux

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func terminalPassword(ctx context.Context, request authPrompt) ([]byte, error) {
	fd, err := unix.Open("/dev/tty", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errors.New("cannot open controlling terminal; use --password-stdin with login or unlock")
	}
	tty := os.NewFile(uintptr(fd), "/dev/tty")
	defer tty.Close()
	return readTerminalPassword(ctx, tty, fd, request)
}

// Raw, nonblocking terminal input lets cancellation restore echo before returning.
func readTerminalPassword(ctx context.Context, tty *os.File, fd int, request authPrompt) (password []byte, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, term.Restore(fd, state))
		_, _ = fmt.Fprint(tty, "\r\n")
		if err != nil {
			clear(password)
			password = nil
		}
	}()
	label := request.Label
	if label == "" {
		label = "Password"
	}
	if _, err := fmt.Fprintf(tty, "%s %q (%q at %q)\r\n%s: ", request.Action, request.Profile, request.Email, request.Server, label); err != nil {
		return nil, err
	}
	buffer := make([]byte, 4096)
	defer func() {
		if password == nil {
			clear(buffer)
		}
	}()
	n := 0
	overlong := false
	var input [1]byte
	// macOS /dev/tty cannot use Go read deadlines. Poll bounds cancellation latency.
	events := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		ready, err := unix.Poll(events, 100)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if ready == 0 {
			continue
		}
		nread, err := unix.Read(fd, input[:])
		if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if nread == 0 {
			return nil, io.EOF
		}

		c := input[0]
		input[0] = 0
		switch c {
		case 3:
			return nil, context.Canceled
		case 4:
			return nil, io.EOF
		case '\r', '\n':
			if n == 0 || overlong {
				return nil, errors.New("invalid password length")
			}
			return buffer[:n], nil
		case 8, 127:
			if n > 0 {
				n--
				for n > 0 && buffer[n]&0xc0 == 0x80 {
					n--
				}
				clear(buffer[n:])
			}
		case 21:
			clear(buffer)
			n = 0
			overlong = false
		default:
			if n == len(buffer) {
				overlong = true
				continue
			}
			buffer[n] = c
			n++
		}
	}
}
