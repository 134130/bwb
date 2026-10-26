// Package daemon wires the vault service into go-daemonkit.
package daemon

import (
	"context"
	"github.com/134130/bwb/internal/config"
	"os"
	"runtime/debug"

	"github.com/134130/bwb/internal/vault"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

var Version string // Set by release builds.

func Options() kit.Options {
	version := Version
	if version == "" {
		version = "dev"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
	return kit.Options{AppID: "bwb", Version: version, InstanceID: os.Getenv("BWB_INSTANCE")}
}
func Manager() (*kit.Manager, error) { return kit.NewManager(Options(), "daemon", "run") }

func Run(ctx context.Context, configFlag string) error {
	path, explicit, err := config.Path(configFlag)
	if err != nil {
		return err
	}
	file, err := config.Load(path, explicit)
	if err != nil {
		return err
	}
	app, err := kit.New(Options())
	if err != nil {
		return err
	}
	store, err := app.Storage("vault")
	if err != nil {
		return err
	}
	secrets, err := app.Secrets("vault")
	if err != nil {
		return err
	}
	service := &vault.Service{Store: store, Secrets: secrets, InitialConfig: file}
	if err := app.Register(service); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.DaemonStatus, rpcHandler(func(context.Context, vault.Empty) (vaultapi.DaemonStatusResponse, error) {
		return vaultapi.DaemonStatusResponse{ConfigPath: path}, nil
	})); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.TOTP, rpcHandler(service.TOTP)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Describe, rpcHandler(service.Describe)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Profiles, rpcHandler(service.Profiles)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Activity, rpcHandler(service.Activity)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Reload, rpcHandler(func(ctx context.Context, _ vault.Empty) (vault.Empty, error) {
		file, err := config.Load(path, true)
		if err != nil {
			return vault.Empty{}, err
		}
		return vault.Empty{}, service.ApplyConfig(ctx, file)
	})); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.WatchAuth, rpcHandler(service.WatchAuth)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.BeginAuth, rpcHandler(service.BeginAuth)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.EndAuth, rpcHandler(service.EndAuth)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Unlock, rpcHandler(service.Unlock)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Logout, rpcHandler(service.Logout)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Login, rpcHandler(service.Login)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.LoginAPIKey, rpcHandler(service.LoginAPIKey)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.List, rpcHandler(service.List)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Read, rpcHandler(service.Read)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Item, rpcHandler(service.Item)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Sync, rpcHandler(service.Sync)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Lock, rpcHandler(service.Lock)); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Status, rpcHandler(service.Status)); err != nil {
		return err
	}
	return app.Run(ctx)
}
