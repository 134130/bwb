// Package vaultapi defines the shared CLI/daemon wire contract.
package vaultapi

import (
	"github.com/134130/bwb/internal/vault"
	daemon "github.com/134130/go-daemonkit"
)

type ListRequest = vault.ListRequest
type ListResponse = vault.ListResponse
type ListEntry = vault.ListEntry
type NamedID = vault.NamedID

type ProfileRequest = vault.ProfileRequest
type LoginRequest = vault.LoginRequest
type UnlockRequest = vault.UnlockRequest
type ReadRequest = vault.ReadRequest
type Empty = vault.Empty
type ReadResponse = vault.ReadResponse
type StatusResponse = vault.StatusResponse

const (
	UnlockRequiredCode = -32010
	LoginRequiredCode  = -32011
	StateChangedCode   = -32012
)

type AuthDetails struct {
	Profile string `json:"profile"`
	Epoch   uint64 `json:"epoch"`
	Version uint64 `json:"version"`
}

type AuthRequest = vault.AuthRequest
type AuthResponse = vault.AuthResponse
type AuthEndRequest = vault.AuthEndRequest
type AuthWatchResponse = vault.AuthWatchResponse

const AuthTimeout = vault.AuthTimeout

var (
	List      = daemon.NewMethod[ListRequest, ListResponse]("vault.list")
	WatchAuth = daemon.NewMethod[AuthEndRequest, AuthWatchResponse]("vault.watchAuth")
	BeginAuth = daemon.NewMethod[AuthRequest, AuthResponse]("vault.beginAuth")
	EndAuth   = daemon.NewMethod[AuthEndRequest, Empty]("vault.endAuth")
	Unlock    = daemon.NewMethod[UnlockRequest, Empty]("vault.unlock")
	Logout    = daemon.NewMethod[ProfileRequest, Empty]("vault.logout")
	Login     = daemon.NewMethod[LoginRequest, Empty]("vault.login")
	Read      = daemon.NewMethod[ReadRequest, ReadResponse]("vault.read")
	Sync      = daemon.NewMethod[ProfileRequest, Empty]("vault.sync")
	Lock      = daemon.NewMethod[ProfileRequest, Empty]("vault.lock")
	Status    = daemon.NewMethod[ProfileRequest, StatusResponse]("vault.status")
)
