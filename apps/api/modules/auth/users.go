// Package auth wires Opus's account table into porte's login. porte verifies
// the credential and hands back an app-side int64 user id; porte/kit's store
// creates and finds the row porte resolved.
package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"
)

// Columns is where Opus's users table differs from the kit's defaults: the
// picture lives in image, and every write the kit makes bumps updated_at.
func Columns() kit.Columns {
	return kit.Columns{Picture: "image", UpdatedAt: "updated_at"}
}

// UserStore is porte/kit's store over Opus's users table, plus the one claim
// the kit does not know: the GitHub account Registre reports under
// connections, which links pull requests to their author.
type UserStore struct {
	*kit.Store
}

var _ porte.UserStore = (*UserStore)(nil)

// NewUserStore builds the store. eraseData deletes what an erased person owns
// before the kit deletes their users row, in the same transaction.
func NewUserStore(ctx context.Context, db *sql.DB, eraseData func(context.Context, *sql.Tx, int64) error) (*UserStore, error) {
	store, err := kit.NewStore(ctx, kit.Store{DB: db, Columns: Columns(), OnCreate: fillNewAccount, EraseData: eraseData})
	if err != nil {
		return nil, err
	}
	return &UserStore{Store: store}, nil
}

// UpsertFromOIDC applies the kit's sign-in rules, then stores the GitHub id
// when the claims carry one.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	userID, err := s.Store.UpsertFromOIDC(ctx, claims)
	if err != nil {
		return 0, err
	}
	githubID := extractGitHubID(claims.Extra)
	if githubID == "" {
		return userID, nil
	}
	if _, err := s.DB.ExecContext(ctx, `UPDATE users SET github_id = $1 WHERE id = $2`, githubID, userID); err != nil {
		return 0, errors.Internal("failed to store the github account", err)
	}
	return userID, nil
}

// fillNewAccount writes the empty colour every account created before the
// kit carries, so a new row reads like the old ones. github_id stays NULL
// until Registre reports a GitHub account, as it always has.
func fillNewAccount(ctx context.Context, tx *sql.Tx, userID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE users SET color = '' WHERE id = $1`, userID)
	return err
}

func extractGitHubID(extra map[string]json.RawMessage) string {
	raw, ok := extra["connections"]
	if !ok || len(raw) == 0 {
		return ""
	}
	var connections struct {
		GitHub string `json:"github"`
	}
	if err := json.Unmarshal(raw, &connections); err != nil {
		return ""
	}
	return strings.TrimSpace(connections.GitHub)
}
