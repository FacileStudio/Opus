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

// workspaceHooks are what the workspace feeds call back into: onSynced and
// onSnapshot for the Registre spaces, onPersonDeleted for an erasure notice.
type workspaceHooks struct {
	onSynced        func(context.Context)
	onSnapshot      func(directory.Snapshot)
	onPersonDeleted func(context.Context, string, string) error
}

// startWorkspaces builds the workspace service and, when OIDC is configured,
// hands workspace management to Registre: it imports the existing workspaces
// once, in the background and retrying until Registre accepts them, then
// keeps them synced. It also starts the GDPR deletion receiver, which polls
// Registre with the same client credentials. Opus runs one replica, so this is
// the one deletion poller. Without OIDC, Opus manages workspaces itself and
// there is no sync route. hooks.onSynced runs once: after the first Registre
// snapshot is applied, or right away without OIDC.
func startWorkspaces(db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger, hooks workspaceHooks) (
	*workspace.Service, registreFollowers, func(),
) {
	workspaces := workspace.NewService(db, appEnv.UploadDir)
	if appEnv.OIDC == nil {
		go hooks.onSynced(context.Background())
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
	spaces.OnFirstSync(hooks.onSynced)
	spaces.OnSnapshot(hooks.onSnapshot)
	deletions := workspace.NewDeletions(workspaces, client, hooks.onPersonDeleted, portepg.New(sqlDB).EraseIdentity, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	go spaces.Run(ctx)
	go deletions.Run(ctx)
	return workspaces, registreFollowers{spaces: spaces, deletions: deletions}, cancel
}
