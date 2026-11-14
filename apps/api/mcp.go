package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/modules/mcp"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"gorm.io/gorm"
)

// newMCP builds the MCP server when OIDC is on and MCP_RESOURCE is set, and
// nothing otherwise. The verifier runs Registre discovery here, so a wrong
// issuer fails at boot like the OIDC kit does. A person's first MCP call
// creates their account through the OIDC kit, as a first web login would.
// Resolved and expired pending actions are pruned until ctx ends.
func newMCP(ctx context.Context, kits *authKits, appEnv *env.Config, tools mcp.Domain, appLogger *slog.Logger) (*mcp.Service, error) {
	if appEnv.OIDC == nil || appEnv.MCP.Resource == "" {
		return nil, nil
	}
	sqlDB, err := tools.DB.DB()
	if err != nil {
		return nil, err
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{
		Issuer:      appEnv.OIDC.Issuer,
		Resource:    appEnv.MCP.Resource,
		ScopePrefix: "opus",
		Identities:  kits.store.Identities(),
		Provision:   true,
		Accounts:    kits.oidc,
		Logger:      appLogger,
	})
	if err != nil {
		return nil, err
	}
	pending := mcpkit.PostgresPending{DB: sqlDB}
	service, err := mcp.NewService(tools, mcp.Config{
		Resource: appEnv.MCP.Resource,
		Issuer:   appEnv.OIDC.Issuer,
		Version:  version,
		Origins:  appEnv.CORSAllowedOrigins,
		StateKey: appEnv.MCP.StateKey,
		Verifier: verifier,
		Pending:  pending,
		Logger:   appLogger,
	})
	if err != nil {
		return nil, err
	}
	go mcpkit.KeepPruned(ctx, pending, 0, 0, appLogger)
	appLogger.Info("mcp server on", slog.String("resource", appEnv.MCP.Resource))
	return service, nil
}

// mcpDomain is the shared services the MCP tools call.
func mcpDomain(db *gorm.DB, workspaces *workspace.Service, shared domain) mcp.Domain {
	return mcp.Domain{
		Workspaces: workspaces,
		Projects:   shared.projects,
		Columns:    shared.columns,
		Tasks:      shared.tasks,
		Labels:     shared.labels,
		Activities: shared.activities,
		DB:         db,
	}
}
