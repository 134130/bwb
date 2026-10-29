// Package vaultapi defines the shared CLI/daemon wire contract.
package vaultapi

import (
	"github.com/134130/bwb/internal/vault"
	daemon "github.com/134130/go-daemonkit"
)

type RepromptRequest = vault.RepromptRequest
type RepromptResponse = vault.RepromptResponse
type RepromptWatchRequest = vault.RepromptWatchRequest
type RepromptDetails = vault.RepromptRequired

type TOTPRequest = vault.TOTPRequest
type TOTPResponse = vault.TOTPResponse

type ListRequest = vault.ListRequest
type ListResponse = vault.ListResponse
type ListEntry = vault.ListEntry
type NamedID = vault.NamedID

type ProfileRequest = vault.ProfileRequest
type LoginRequest = vault.LoginRequest
type LoginAPIKeyRequest = vault.LoginAPIKeyRequest
type LoginSSORequest = vault.LoginSSORequest
type LoginPlanRequest = vault.LoginPlanRequest
type LoginPlanResponse = vault.LoginPlanResponse
type ExchangeSSORequest = vault.ExchangeSSORequest
type FinishSSOResponse = vault.FinishSSOResponse
type UnlockRequest = vault.UnlockRequest
type ReadRequest = vault.ReadRequest
type Empty = vault.Empty
type ReadResponse = vault.ReadResponse
type ItemRequest = vault.ItemRequest
type ItemResponse = vault.ItemResponse
type DescribeRequest = vault.DescribeRequest
type DescribeResponse = vault.DescribeResponse
type ProfilesResponse = vault.ProfilesResponse

type StatusResponse = vault.StatusResponse

type DaemonStatusResponse struct {
	ConfigOptional bool   `json:"configOptional"`
	ConfigPath     string `json:"configPath"`
}

const (
	RepromptRequiredCode = -32013
	UnlockRequiredCode   = -32010
	LoginRequiredCode    = -32011
	StateChangedCode     = -32012
)

type AuthDetails struct {
	Profile string `json:"profile"`
	Method  string `json:"method,omitempty"`
	Epoch   uint64 `json:"epoch"`
	Version uint64 `json:"version"`
}

type AuthRequest = vault.AuthRequest
type AuthResponse = vault.AuthResponse
type AuthEndRequest = vault.AuthEndRequest
type AuthWatchResponse = vault.AuthWatchResponse

const AuthTimeout = vault.AuthTimeout
const SSOTimeout = vault.SSOTimeout

var (
	Reprompt      = daemon.NewMethod[RepromptRequest, RepromptResponse]("vault.reprompt")
	WatchReprompt = daemon.NewMethod[RepromptWatchRequest, Empty]("vault.watchReprompt")
	TOTP          = daemon.NewMethod[TOTPRequest, TOTPResponse]("vault.totp")
	Describe      = daemon.NewMethod[DescribeRequest, DescribeResponse]("vault.describe")
	Profiles      = daemon.NewMethod[Empty, ProfilesResponse]("vault.profiles")
	Activity      = daemon.NewMethod[ProfileRequest, Empty]("vault.activity")
	DaemonStatus  = daemon.NewMethod[Empty, DaemonStatusResponse]("daemon.status")
	Reload        = daemon.NewMethod[Empty, Empty]("vault.reload")
	List          = daemon.NewMethod[ListRequest, ListResponse]("vault.list")
	WatchAuth     = daemon.NewMethod[AuthEndRequest, AuthWatchResponse]("vault.watchAuth")
	BeginAuth     = daemon.NewMethod[AuthRequest, AuthResponse]("vault.beginAuth")
	EndAuth       = daemon.NewMethod[AuthEndRequest, Empty]("vault.endAuth")
	Unlock        = daemon.NewMethod[UnlockRequest, Empty]("vault.unlock")
	Logout        = daemon.NewMethod[ProfileRequest, Empty]("vault.logout")
	Login         = daemon.NewMethod[LoginRequest, Empty]("vault.login")
	LoginAPIKey   = daemon.NewMethod[LoginAPIKeyRequest, Empty]("vault.loginApiKey")
	LoginSSO      = daemon.NewMethod[LoginSSORequest, Empty]("vault.loginSSO")
	PlanLogin     = daemon.NewMethod[LoginPlanRequest, LoginPlanResponse]("vault.planLogin")
	ExchangeSSO   = daemon.NewMethod[ExchangeSSORequest, AuthResponse]("vault.exchangeSSO")
	FinishSSO     = daemon.NewMethod[UnlockRequest, FinishSSOResponse]("vault.finishSSO")
	Read          = daemon.NewMethod[ReadRequest, ReadResponse]("vault.read")
	Item          = daemon.NewMethod[ItemRequest, ItemResponse]("vault.item")
	Sync          = daemon.NewMethod[ProfileRequest, Empty]("vault.sync")
	Lock          = daemon.NewMethod[ProfileRequest, Empty]("vault.lock")
	Status        = daemon.NewMethod[ProfileRequest, StatusResponse]("vault.status")
)
