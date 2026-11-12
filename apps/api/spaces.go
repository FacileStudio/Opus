package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/modules/auth"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/porte"
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
// onSnapshot for the Registre spaces, and the deletion receiver, which the
// space sync also asks before it provisions a member.
type workspaceHooks struct {
	onSynced   func(context.Context)
	onSnapshot func(directory.Snapshot)
	deletions  *workspace.Deletions
}

// newDeletions builds the GDPR deletion receiver when OIDC is configured, and
// nil otherwise. It exists before porte's OIDC kit, which refuses an account
// to every subject it handled.
func newDeletions(db *gorm.DB, appEnv *env.Config, users *auth.UserStore, store *portepg.Store, appLogger *slog.Logger) *workspace.Deletions {
	if appEnv.OIDC == nil {
		return nil
	}
	return workspace.NewDeletions(db, registreClient(appEnv.OIDC), users.EraseAccount, store, appLogger)
}

// erasedSubjects is the receiver as porte/oidc's guard, or no guard without one.
func erasedSubjects(deletions *workspace.Deletions) porte.ErasedSubjects {
	if deletions == nil {
		return nil
	}
	return deletions
}

func registreClient(cfg *env.OIDCConfig) directory.Client {
	return directory.Client{Issuer: cfg.Issuer, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret}
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
	if appEnv.OIDC == nil || hooks.deletions == nil {
		go hooks.onSynced(context.Background())
		return workspaces, registreFollowers{}, func() {}
	}
	spaces := workspace.NewDirectory(workspaces, registreClient(appEnv.OIDC), appLogger)
	spaces.OnFirstSync(hooks.onSynced)
	spaces.OnSnapshot(hooks.onSnapshot)
	spaces.RefuseErased(hooks.deletions)
	ctx, cancel := context.WithCancel(context.Background())
	go spaces.Run(ctx)
	go hooks.deletions.Run(ctx)
	return workspaces, registreFollowers{spaces: spaces, deletions: hooks.deletions}, cancel
}
