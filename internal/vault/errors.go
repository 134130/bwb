package vault

import "errors"

var ErrStateChanged = errors.New("vault state changed; retry the command")

// AuthRequired describes recoverable state, not rejected credentials or infrastructure failure.
type AuthRequired struct {
	Action  string
	Profile string
	Epoch   uint64
	Version uint64
}

func (e *AuthRequired) Error() string {
	if e.Action == "unlock" {
		return "vault locked; run unlock"
	}
	return "login required; run login"
}
