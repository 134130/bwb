package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/134130/bwb/internal/vault"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
	"github.com/134130/go-daemonkit/secrets"
)

func TestDomainErrorTranslation(t *testing.T) {
	for _, action := range []string{"login", "unlock"} {
		handler := rpcHandler(func(context.Context, vault.Empty) (vault.ReadResponse, error) {
			return vault.ReadResponse{Value: "must not escape"}, &vault.AuthRequired{Action: action, Method: "apikey", Profile: "work", Epoch: 4, Version: 7}
		})
		result, err := handler(context.Background(), vault.Empty{})
		var rpc *kit.RPCError
		if result.Value != "" || !errors.As(err, &rpc) {
			t.Fatal("invalid translated response", err)
		}
		want := vaultapi.UnlockRequiredCode
		if action == "login" {
			want = vaultapi.LoginRequiredCode
		}
		var detail vaultapi.AuthDetails
		if rpc.Code != want || json.Unmarshal(rpc.Data, &detail) != nil || detail.Profile != "work" || detail.Method != "apikey" || detail.Epoch != 4 || detail.Version != 7 {
			t.Fatal("lost auth contract")
		}
	}
	for _, failure := range []error{bitwarden.ErrAuthentication, bitwarden.ErrCipher, secrets.ErrPermission, errors.New("network unavailable"), context.Canceled} {
		_, err := rpcHandler(func(context.Context, vault.Empty) (vault.Empty, error) { return vault.Empty{}, failure })(context.Background(), vault.Empty{})
		var rpc *kit.RPCError
		if !errors.Is(err, failure) || errors.As(err, &rpc) {
			t.Fatal("ordinary failure became authentication request", err)
		}
	}
}

func TestRepromptRPCContract(t *testing.T) {
	result, err := rpcHandler(func(context.Context, vault.Empty) (vault.ReadResponse, error) {
		return vault.ReadResponse{Value: "secret"}, &vault.RepromptRequired{Profile: "work", Item: "item-id"}
	})(t.Context(), vault.Empty{})
	var rpc *kit.RPCError
	if result.Value != "" || !errors.As(err, &rpc) || rpc.Code != vaultapi.RepromptRequiredCode {
		t.Fatal("missing reprompt code", err)
	}
	var details vaultapi.RepromptDetails
	if json.Unmarshal(rpc.Data, &details) != nil || details.Profile != "work" || details.Item != "item-id" {
		t.Fatal("missing reprompt identity")
	}
}
