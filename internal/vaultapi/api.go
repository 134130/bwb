// Package vaultapi defines the shared CLI/daemon wire contract.
package vaultapi

import daemon "github.com/134130/go-daemonkit"

type ProfileRequest struct {
	Profile string `json:"profile"`
}
type LoginRequest struct {
	Profile  string `json:"profile"`
	Server   string `json:"server"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}
type UnlockRequest struct {
	Profile  string `json:"profile"`
	Password []byte `json:"password"`
}

type ReadRequest struct {
	Profile string `json:"profile"`
	ID      string `json:"id"`
	Field   string `json:"field"`
}
type Empty struct{}
type ReadResponse struct {
	Value string `json:"value"`
}
type StatusResponse struct {
	State string `json:"state"`
}

const LockedCode = -32010

var (
	Unlock = daemon.NewMethod[UnlockRequest, Empty]("vault.unlock")
	Logout = daemon.NewMethod[ProfileRequest, Empty]("vault.logout")
	Login  = daemon.NewMethod[LoginRequest, Empty]("vault.login")
	Read   = daemon.NewMethod[ReadRequest, ReadResponse]("vault.read")
	Sync   = daemon.NewMethod[ProfileRequest, Empty]("vault.sync")
	Lock   = daemon.NewMethod[ProfileRequest, Empty]("vault.lock")
	Status = daemon.NewMethod[ProfileRequest, StatusResponse]("vault.status")
)
