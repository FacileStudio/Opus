package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/porte/forfait"
	"github.com/FacileStudio/tronc/bulle"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const (
	ideasApp         = "opus"
	aiAccessKey      = "ai.access"
	bulleAudience    = "bulle"
	workspaceIDParam = "workspaceId"
)

type ideas struct {
	db         *gorm.DB
	kits       *authKits
	env        *env.Config
	logger     *slog.Logger
	plans      *forfait.Store
	exchanger  *exchange.Exchanger
	configured bool
}

func newIdeas(db *gorm.DB, kits *authKits, appEnv *env.Config, logger *slog.Logger) *ideas {
	i := &ideas{db: db, kits: kits, env: appEnv, logger: logger}
	if appEnv.OIDC == nil || appEnv.BulleURL == "" {
		return i
	}
	i.plans = forfait.NewStore(forfait.Feature{Key: aiAccessKey, Kind: forfait.Bool, Default: []byte("false")})
	i.plans.Issuer = appEnv.OIDC.Issuer
	exchanger, err := exchange.New(exchange.Config{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
		Identities:   kits.store.Identities(),
		Logger:       logger,
	})
	if err != nil {
		logger.Error("ideas disabled", slog.Any("error", err))
		return i
	}
	i.exchanger = exchanger
	i.configured = true
	return i
}

func (i *ideas) apply(snapshot directory.Snapshot) {
	if i.plans != nil {
		i.plans.Apply(snapshot)
	}
}

func (i *ideas) mount(api chi.Router, authService middleware.Authenticator) {
	if !i.configured {
		return
	}
	source := i.exchanger.Source(exchange.Target{
		Audience: bulleAudience,
		Scopes:   []string{"bulle.read", "bulle.write"},
	})
	api.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authService))
		err := bulle.MountIdeas(r, bulle.Options{
			App:         ideasApp,
			URL:         i.env.BulleURL,
			TokenSource: source.Token,
			Forget:      source.Forget,
			Space:       i.space,
			AIAccess:    i.aiAccess,
			UpgradeURL:  strings.TrimSuffix(i.env.OIDC.Issuer, "/") + "/spaces",
		})
		if err != nil {
			i.logger.Error("ideas not mounted", slog.Any("error", err))
		}
	})
}

func (i *ideas) space(r *http.Request) string {
	identity, ok := porte.From(r.Context())
	if !ok {
		return ""
	}
	workspaceID, err := strconv.ParseInt(r.URL.Query().Get(workspaceIDParam), 10, 64)
	if err != nil || workspaceID <= 0 {
		return ""
	}
	var spaceID *string
	err = i.db.WithContext(r.Context()).
		Model(&schemas.Workspace{}).
		Select("workspaces.registre_space_id").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspaces.id = ? AND workspace_members.user_id = ?", workspaceID, identity.UserID).
		Limit(1).
		Scan(&spaceID).Error
	if err != nil || spaceID == nil {
		return ""
	}
	return *spaceID
}

func (i *ideas) aiAccess(r *http.Request) bool {
	if spaceID := i.space(r); spaceID != "" {
		return i.plans.Enabled(forfait.Scope{SpaceID: spaceID}, aiAccessKey)
	}
	identity, ok := porte.From(r.Context())
	if !ok {
		return false
	}
	stored, err := i.kits.store.Identities().ListByUser(r.Context(), identity.UserID)
	if err != nil {
		return false
	}
	issuer := i.env.OIDC.Issuer
	for _, entry := range stored {
		if entry.Provider == issuer && i.plans.Enabled(forfait.Scope{Subject: entry.Subject}, aiAccessKey) {
			return true
		}
	}
	return false
}
