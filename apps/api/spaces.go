package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

// registreFollowers are the two Registre feeds Opus follows. Both are nil
// without OIDC.
type registreFollowers struct {
	spaces    *workspace.Directory
	deletions *workspace.Deletions
}

// startWorkspaces builds the workspace service and, when OIDC is configured,
// hands workspace management to Registre: it imports the existing workspaces
// once, in the background and retrying until Registre accepts them, then
// keeps them synced. It also starts the GDPR deletion receiver, which polls
// Registre with the same client credentials. Opus runs one replica, so this is
// the one deletion poller. Without OIDC, Opus manages workspaces itself and
// there is no sync route. onSynced runs once: after the first Registre
// snapshot is applied, or right away without OIDC.
func startWorkspaces(db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger, onSynced func(context.Context)) (
	*workspace.Service, registreFollowers, func(),
) {
	workspaces := workspace.NewService(db, appEnv.UploadDir)
	if appEnv.OIDC == nil {
		go onSynced(context.Background())
		return workspaces, registreFollowers{}, func() {}
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("deletion receiver not started", slog.Any("error", err))
		return workspaces, registreFollowers{}, func() {}
	}
	client := directory.Client{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
	}
	spaces := workspace.NewDirectory(workspaces, client, appLogger)
	spaces.OnFirstSync(onSynced)
	deletions := workspace.NewDeletions(workspaces, client, portepg.New(sqlDB).EraseIdentity, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	go spaces.Run(ctx)
	go deletions.Run(ctx)
	return workspaces, registreFollowers{spaces: spaces, deletions: deletions}, cancel
}
