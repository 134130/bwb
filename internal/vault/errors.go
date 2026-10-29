package vault

import "errors"

var ErrStateChanged = errors.New("vault state changed; retry the command")

// AuthRequired describes recoverable state, not rejected credentials or infrastructure failure.
type AuthRequired struct {
	Action  string
	Method  string
	Profile string
	Epoch   uint64
	Version uint64
}

func (e *AuthRequired) Error() string {
	if e.Action == "unlock" {
		return "vault locked; run unlock"
	}
	if e.Method == "apikey" {
		return "profile uses API key login; run login --apikey, or login --sso for browser SSO"
	}
	if e.Method == "sso" {
		return "SSO login required; run login --sso"
	}
	return "login required; run login"
}
