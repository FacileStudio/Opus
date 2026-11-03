package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// startWorkspaces builds the workspace service and, when OIDC is configured,
// hands workspace management to Registre: it imports the existing workspaces
// once, in the background and retrying until Registre accepts them, then
// keeps them synced. Without OIDC, Opus manages workspaces itself and there is
// no sync route. onSynced runs once: after the first Registre snapshot is
// applied, or right away without OIDC.
func startWorkspaces(db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger, onSynced func(context.Context)) (
	*workspace.Service, *workspace.Directory, func(),
) {
	workspaces := workspace.NewService(db, appEnv.UploadDir)
	if appEnv.OIDC == nil {
		go onSynced(context.Background())
		return workspaces, nil, func() {}
	}
	spaces := workspace.NewDirectory(workspaces, directory.Client{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
	}, appLogger)
	spaces.OnFirstSync(onSynced)
	ctx, cancel := context.WithCancel(context.Background())
	go spaces.Run(ctx)
	return workspaces, spaces, cancel
}
