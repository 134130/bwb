package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/synctest"
)

func TestHelpAndInputErrors(t *testing.T) {
	isolateConfig(t)
	root := NewRoot()
	var help bytes.Buffer
	root.SetOut(&help)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil || !strings.Contains(help.String(), "read") {
		t.Fatal("Cobra help failed", err)
	}
	for _, args := range [][]string{
		{"read"},
		{"read", "bad", "login/password"},
		{"login"},
		{"login", "--server", "https://example.invalid", "--email", "user@example.invalid"},
	} {
		root := NewRoot()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&out)
		root.SetIn(strings.NewReader(""))
		root.SetArgs(args)
		if err := root.Execute(); err == nil {
			t.Fatal("invalid command succeeded")
		}
		if out.Len() != 0 {
			t.Fatal("command error wrote unexpected output")
		}
	}
}

func TestPasswordInputBounded(t *testing.T) {
	isolateConfig(t)
	for _, input := range []string{"", "\n", strings.Repeat("x", 4099), strings.Repeat("x", 4096) + "\r\n" + "ignored"} {
		root := NewRoot()
		root.SetIn(strings.NewReader(input))
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"login", "--server", "http://127.0.0.1:1", "--email", "user@example.invalid", "--password-stdin"})
		err := root.Execute()
		if err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatal("invalid password input reached network", err)
		}
	}
}

func TestPasswordInputCancellation(t *testing.T) {
	isolateConfig(t)
	synctest.Test(t, func(t *testing.T) {
		reader, writer := io.Pipe()
		defer reader.Close()
		defer writer.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		root := NewRoot()
		root.SetIn(reader)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"login", "--server", "http://127.0.0.1:1", "--email", "user@example.invalid", "--password-stdin"})
		result := make(chan error, 1)
		go func() { result <- root.ExecuteContext(ctx) }()
		if _, err := writer.Write([]byte("partial password")); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		cancel()
		synctest.Wait()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected cancellation, got %v", err)
			}
		default:
			t.Fatal("password input remained blocked after cancellation")
		}
	})
}

func TestPasswordReadCompletesWithoutClosingInput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		reader, writer := io.Pipe()
		defer reader.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			defer writer.Close()
			_, _ = writer.Write([]byte("password\r\n"))
		}()
		password, err := readPassword(ctx, reader)
		if err != nil || string(password) != "password\r\n" {
			t.Fatal("password read failed", err)
		}
		clear(password)
		cancel()
		synctest.Wait()
		if _, err := reader.Read(make([]byte, 1)); err != io.EOF {
			t.Fatal("reader closed after completed read", err)
		}
	})
}

type blockingCloseReader struct {
	*io.PipeReader
	release <-chan struct{}
}

func (r blockingCloseReader) Close() error {
	<-r.release
	return r.PipeReader.Close()
}

func TestPasswordCancellationDoesNotWaitForClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		reader, writer := io.Pipe()
		defer reader.Close()
		defer writer.Close()
		release := make(chan struct{})
		defer close(release)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result := make(chan error, 1)
		go func() {
			_, err := readPassword(ctx, blockingCloseReader{reader, release})
			result <- err
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		default:
			t.Fatal("cancellation waited for blocked Close")
		}
	})
}
