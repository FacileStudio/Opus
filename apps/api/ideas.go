package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/porte/forfait"
	"github.com/FacileStudio/tronc/bulle"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
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
	synced     atomic.Bool
}

func newIdeas(db *gorm.DB, kits *authKits, appEnv *env.Config, logger *slog.Logger) *ideas {
	i := &ideas{db: db, kits: kits, env: appEnv, logger: logger}
	if appEnv.OIDC == nil || appEnv.BulleURL == "" {
		return i
	}
	i.plans = forfait.NewStore(forfait.Feature{Key: aiAccessKey, Kind: forfait.Bool, Default: []byte("false")})
	i.plans.Issuer = appEnv.OIDC.Issuer
	i.plans.Logger = logger
	i.loadPlans(context.Background())
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
		i.savePlans(snapshot)
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
		r.Use(middleware.RequireAuth(authService), i.resolve)
		err := bulle.MountIdeas(r, bulle.Options{
			App:         ideasApp,
			URL:         i.env.BulleURL,
			TokenSource: source.Token,
			Forget:      source.Forget,
			Space:       activeSpace,
			AIAccess:    i.aiAccess,
			UpgradeURL:  strings.TrimSuffix(i.env.OIDC.Issuer, "/") + "/spaces",
		})
		if err != nil {
			i.logger.Error("ideas not mounted", slog.Any("error", err))
		}
	})
}

// spaceKey holds, in a request's context, the Registre space resolved from
// its workspaceId: empty for the person's own scope.
type spaceKey struct{}

// resolve answers the error of an unusable workspaceId, and otherwise passes
// the request on with its Registre space resolved once.
func (i *ideas) resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scoped, err := i.scoped(r)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, scoped)
	})
}

// scoped is r carrying the Registre space of its workspaceId.
func (i *ideas) scoped(r *http.Request) (*http.Request, error) {
	spaceID, err := i.space(r)
	if err != nil {
		return nil, err
	}
	return r.WithContext(context.WithValue(r.Context(), spaceKey{}, spaceID)), nil
}

// activeSpace is the Registre space resolve put on r, empty for personal.
func activeSpace(r *http.Request) string {
	spaceID, _ := r.Context().Value(spaceKey{}).(string)
	return spaceID
}

// space maps the workspaceId query to the Registre space it mirrors. No
// workspaceId, or the caller's personal workspace, is the personal scope.
// A workspace the caller is not a member of is not found, and a shared one
// not linked to Registre yet is refused: neither falls back to personal.
func (i *ideas) space(r *http.Request) (string, error) {
	identity, ok := porte.From(r.Context())
	if !ok {
		return "", errors.Unauthorized("missing auth token")
	}
	raw := r.URL.Query().Get(workspaceIDParam)
	if raw == "" {
		return "", nil
	}
	workspaceID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || workspaceID <= 0 {
		return "", errors.Invalid("invalid workspaceId")
	}
	var row workspaceLink
	found := i.db.WithContext(r.Context()).
		Model(&schemas.Workspace{}).
		Select("workspaces.registre_space_id, workspaces.personal_owner_id").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspaces.id = ? AND workspace_members.user_id = ?", workspaceID, identity.UserID).
		Limit(1).
		Scan(&row)
	if found.Error != nil {
		return "", errors.Internal("failed to load the workspace", found.Error)
	}
	return row.space(found.RowsAffected > 0)
}

// workspaceLink is what space reads of a workspace.
type workspaceLink struct {
	RegistreSpaceID *string
	PersonalOwnerID *int64
}

// space is the Registre space of a workspace the caller is a member of
// (found), or why it has none.
func (w workspaceLink) space(found bool) (string, error) {
	switch {
	case !found:
		return "", errors.NotFound("workspace not found")
	case w.PersonalOwnerID != nil:
		return "", nil
	case w.RegistreSpaceID == nil || *w.RegistreSpaceID == "":
		return "", errors.Failed("this workspace is not linked to Registre yet, so Bulle cannot suggest ideas for it")
	}
	return *w.RegistreSpaceID, nil
}

// aiAccess reports whether the active space's plan, or the person's own,
// includes ai.access. It fails closed: until Registre has delivered
// entitlements once, AI is off, even though forfait itself fails open.
func (i *ideas) aiAccess(r *http.Request) bool {
	if !i.synced.Load() {
		return false
	}
	if spaceID := activeSpace(r); spaceID != "" {
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
