// Package daemon wires the vault service into go-daemonkit.
package daemon

import (
	"context"
	"os"

	"github.com/134130/bwb/internal/vault"
	"github.com/134130/bwb/internal/vaultapi"
	kit "github.com/134130/go-daemonkit"
)

func Options() kit.Options {
	return kit.Options{AppID: "bwb", Version: "dev", InstanceID: os.Getenv("BWB_INSTANCE")}
}
func Manager() (*kit.Manager, error) { return kit.NewManager(Options(), "daemon", "run") }

func Connect(ctx context.Context) (*kit.Client, error) {
	manager, err := Manager()
	if err != nil {
		return nil, err
	}
	if err := manager.EnsureRunning(ctx); err != nil {
		return nil, err
	}
	return kit.Connect(ctx, Options())
}

func Run(ctx context.Context) error {
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
	service := &vault.Service{Store: store, Secrets: secrets}
	if err := app.Register(service); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Unlock, service.Unlock); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Logout, service.Logout); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Login, service.Login); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Read, service.Read); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Sync, service.Sync); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Lock, service.Lock); err != nil {
		return err
	}
	if err := app.Handle(vaultapi.Status, service.Status); err != nil {
		return err
	}
	return app.Run(ctx)
}
