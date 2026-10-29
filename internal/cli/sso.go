package cli

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/134130/bwb/internal/vaultclient"
	kit "github.com/134130/go-daemonkit"
	"github.com/awnumar/memguard"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func commandLogin(cmd *cobra.Command, profile, server, email, identifier string, sso, stdin, nonInteractive bool) error {
	if sso && nonInteractive {
		return errors.New("SSO login requires browser interaction; remove --non-interactive")
	}
	if !stdin && interactivePrompt(cmd, nonInteractive) == nil {
		return errors.New("--password-stdin is required in non-interactive mode")
	}
	if cmd.Flags().Changed("identifier") && strings.TrimSpace(identifier) == "" {
		return errors.New("SSO identifier cannot be empty")
	}
	if server != "" {
		if _, err := bitwarden.NewClient(server, email, uuid.NewString()); err != nil {
			return err
		}
	}
	client, err := connectCommand(cmd)
	if err != nil {
		return err
	}
	defer client.Close()
	server, email, err = loginIdentity(cmd, client, profile, server, email, !sso, nonInteractive)
	if err != nil {
		return err
	}
	plan, err := client.Invoke(cmd.Context(), vaultapi.PlanLogin, vaultapi.LoginPlanRequest{Profile: profile, Server: server, Email: email, SSO: sso, Identifier: identifier})
	var rpc *kit.RPCError
	if errors.As(err, &rpc) && rpc.Code == -32601 {
		return errors.New("login discovery requires an updated daemon; run daemon update")
	}
	if err != nil {
		return err
	}
	if plan.Method == "sso" {
		if nonInteractive {
			return errors.New("SSO login requires browser interaction; remove --non-interactive")
		}
		return commandSSO(cmd, client, profile, plan, stdin)
	}
	password, err := commandPassword(cmd, stdin, nonInteractive, authPrompt{Action: "login", Profile: profile, Server: plan.Server, Email: plan.Email})
	defer clear(password)
	if err != nil {
		return err
	}
	_, err = client.Invoke(cmd.Context(), vaultapi.Login, vaultapi.LoginRequest{Profile: profile, Server: plan.Server, Email: plan.Email, Password: password})
	return err
}

func commandSSO(cmd *cobra.Command, client *kit.Client, profile string, plan vaultapi.LoginPlanResponse, stdin bool) error {
	reservation, err := client.Invoke(cmd.Context(), vaultapi.BeginAuth, vaultapi.AuthRequest{Profile: profile, Action: "sso", Epoch: plan.Epoch, Version: plan.Version,
		Server: plan.Server, Email: plan.Email, Identifier: plan.Identifier})
	if err != nil || reservation.Ready {
		return err
	}
	end := vaultapi.AuthEndRequest{Profile: profile, ID: reservation.ID}
	return vaultclient.RunAuth(cmd.Context(), client, end, vaultapi.SSOTimeout, func(ctx context.Context) error {
		password, err := ssoPassword(ctx, cmd, stdin, authPrompt{Action: "login", Profile: profile, Server: reservation.Server, Email: reservation.Email})
		if err != nil {
			return err
		}
		defer password.Destroy()
		code, err := browserSSO(ctx, reservation.Server, reservation.Email, reservation.Identifier, func(ctx context.Context, uri string) error {
			if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Complete SSO login in your browser:\n%s\n", uri); err != nil {
				return err
			}
			program := "xdg-open"
			if runtime.GOOS == "darwin" {
				program = "open"
			}
			if err := exec.CommandContext(ctx, program, uri).Run(); err != nil {
				_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Could not open the browser. Open the URL above on this computer.")
				return err
			}
			return nil
		})
		defer clear(code.Code)
		defer clear(code.Verifier)
		if err != nil {
			return err
		}
		account, err := client.Invoke(ctx, vaultapi.ExchangeSSO, vaultapi.ExchangeSSORequest{AuthEndRequest: end, SSOCode: code})
		clear(code.Code)
		clear(code.Verifier)
		if err != nil {
			return err
		}
		return finishSSO(ctx, cmd, client, end, reservation.Server, account.Email, stdin, password)
	})
}

func ssoPassword(ctx context.Context, cmd *cobra.Command, stdin bool, request authPrompt) (*memguard.LockedBuffer, error) {
	var password []byte
	var err error
	if stdin {
		password, err = readPassword(ctx, cmd.InOrStdin())
		if err == nil {
			password, err = trimPassword(password)
		}
	} else if prompt := interactivePrompt(cmd, false); prompt != nil {
		password, err = prompt(ctx, request)
	} else {
		err = errors.New("--password-stdin is required in non-interactive mode")
	}
	defer clear(password)
	if err != nil {
		return nil, err
	}
	return memguard.NewBufferFromBytes(password), nil
}

func finishSSO(ctx context.Context, cmd *cobra.Command, client *kit.Client, end vaultapi.AuthEndRequest, server, email string, stdin bool, password *memguard.LockedBuffer) error {
	defer func() { password.Destroy() }()
	for {
		result, err := client.Invoke(ctx, vaultapi.FinishSSO, vaultapi.UnlockRequest{Profile: end.Profile, AuthID: end.ID, Password: password.Bytes()})
		password.Destroy()
		if err != nil || !result.RetryPassword {
			return err
		}
		if stdin {
			return errors.New("incorrect master password")
		}
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "Incorrect master password. Try again; browser authentication is still valid."); err != nil {
			return err
		}
		next, err := ssoPassword(ctx, cmd, false, authPrompt{Action: "unlock", Profile: end.Profile, Server: server, Email: email})
		if err != nil {
			return err
		}
		password = next
	}
}

type ssoResult struct {
	code       []byte
	identifier string
	err        error
}

func ssoCallback(state string, result chan<- ssoResult) http.Handler {
	var received atomic.Bool
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		query, err := url.ParseQuery(r.URL.RawQuery)
		// The Bitwarden web vault appends the organization identifier to the original state.
		nonce, identifier, _ := strings.Cut(query.Get("state"), "_identifier=")
		if r.Method != http.MethodGet || r.URL.Path != "/" || err != nil || len(query["state"]) != 1 || nonce != state || bitwarden.ValidateSSOIdentifier(identifier) != nil {
			http.Error(w, "Invalid SSO callback.", http.StatusBadRequest)
			return
		}
		response := ssoResult{identifier: identifier}
		if len(query["error"]) == 1 && query.Get("error") != "" && !query.Has("code") {
			response.err = errors.New("SSO authorization failed or was canceled; retry login --sso")
		} else if len(query["code"]) == 1 && len(query.Get("code")) > 0 && len(query.Get("code")) <= 4096 && !query.Has("error") {
			response.code = []byte(query.Get("code"))
		} else {
			http.Error(w, "Invalid SSO response.", http.StatusBadRequest)
			return
		}
		if !received.CompareAndSwap(false, true) {
			clear(response.code)
			http.Error(w, "SSO response already received.", http.StatusConflict)
			return
		}
		result <- response
		_, _ = fmt.Fprintln(w, "SSO response received. Return to the terminal to finish logging in.")
	})
}

func browserSSO(ctx context.Context, server, email, identifier string, open func(context.Context, string) error) (bitwarden.SSOCode, error) {
	if _, err := bitwarden.NewClient(server, email, uuid.NewString()); err != nil {
		return bitwarden.SSOCode{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var listener net.Listener
	var err error
	for port := 8065; port <= 8070; port++ {
		listener, err = (&net.ListenConfig{}).Listen(ctx, "tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
	}
	if err != nil {
		return bitwarden.SSOCode{}, errors.New("cannot listen for SSO on localhost ports 8065–8070")
	}
	defer listener.Close()
	verifier := make([]byte, 32)
	_, _ = rand.Read(verifier)
	code := bitwarden.SSOCode{Verifier: []byte(base64.RawURLEncoding.EncodeToString(verifier)),
		RedirectURI: fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)}
	clear(verifier)
	defer clear(code.Verifier)
	challenge := sha256.Sum256(code.Verifier)
	state := rand.Text()
	query := url.Values{"clientId": {"cli"}, "redirectUri": {code.RedirectURI}, "state": {state},
		"codeChallenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}}
	if email != "" {
		query.Set("email", email)
	}
	if identifier != "" {
		query.Set("identifier", identifier)
	}
	result := make(chan ssoResult, 1)
	callback := &http.Server{Handler: ssoCallback(state, result), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, MaxHeaderBytes: 16 << 10}
	defer func() {
		shutdown, stop := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer stop()
		_ = callback.Shutdown(shutdown)
		_ = callback.Close()
		select {
		case pending := <-result:
			clear(pending.code)
		default:
		}
	}()
	serveErr := make(chan error, 1)
	go func() { serveErr <- callback.Serve(listener) }()
	if err := open(ctx, strings.TrimRight(server, "/")+"/#/sso?"+query.Encode()); err != nil {
		return bitwarden.SSOCode{}, err
	}
	select {
	case response := <-result:
		if err := ctx.Err(); err != nil {
			clear(response.code)
			return bitwarden.SSOCode{}, err
		}
		if response.err != nil {
			return bitwarden.SSOCode{}, response.err
		}
		code.Code = response.code
		code.Identifier = response.identifier
		if code.Identifier == "" {
			code.Identifier = identifier
		}
		code.Verifier = append([]byte(nil), code.Verifier...)
		return code, nil
	case <-ctx.Done():
		return bitwarden.SSOCode{}, ctx.Err()
	case <-serveErr:
		return bitwarden.SSOCode{}, errors.New("SSO callback listener stopped")
	}
}
