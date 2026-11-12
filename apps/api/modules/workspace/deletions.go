package workspace

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/storage"

	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"
	mcpkit "github.com/FacileStudio/tronc/mcp"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	receiver *directory.Deletions
}

// AssetRemover deletes stored objects by key and treats a missing one as done.
type AssetRemover interface {
	Delete(ctx context.Context, keys []string) error
}

// NewDeletions builds the receiver. onPersonDeleted erases the Opus account
// behind a Registre identity, the user store's EraseAccount; store erases
// porte's own rows and tells a handled subject an account holds again, so the
// next notice erases that account instead of skipping it.
func NewDeletions(
	orm *gorm.DB, client directory.Client, onPersonDeleted func(context.Context, string, string) error,
	store *portepg.Store, logger *slog.Logger,
) *Deletions {
	return &Deletions{receiver: &directory.Deletions{
		Client:          client,
		OnPersonDeleted: onPersonDeleted,
		Erase:           store.EraseIdentity,
		Holds:           store.HoldsIdentity,
		State:           deletionState{orm: orm},
		Logger:          logger,
	}}
}

// Erased reports whether the receiver handled this subject of the issuer:
// porte/oidc's Deps.Erased, and the space sync's guard. It reads the saved
// state on every call, so it needs no running receiver.
func (d *Deletions) Erased(ctx context.Context, provider, subject string) (bool, error) {
	return d.receiver.Erased(ctx, provider, subject)
}

// Run polls until ctx ends. Start it in exactly one replica.
func (d *Deletions) Run(ctx context.Context) {
	d.receiver.Run(ctx)
}

// Handler verifies and applies Registre's signed user.deleted push.
func (d *Deletions) Handler() http.Handler {
	return d.receiver.Handler()
}

const (
	selectPersonalAssets = `SELECT a.object_key FROM asset a
		JOIN workspaces w ON w.id = a.workspace_id WHERE w.personal_owner_id = $1`
	selectPersonalWorkspace = `SELECT id FROM workspaces WHERE personal_owner_id = $1`
	deleteInvitationsToThem = `DELETE FROM invitations WHERE lower(email) = (SELECT lower(email) FROM users WHERE id = $1)`
	forgetPullAuthor        = `UPDATE task_pull_requests SET author_github_id = NULL
		WHERE author_github_id = (SELECT github_id FROM users WHERE id = $1)`
	deleteEventsNamingThem = `DELETE FROM antenne_outbox o WHERE
		strpos(lower(o.payload::text), (SELECT lower(email) FROM users WHERE id = $1)) > 0
		OR EXISTS (SELECT 1 FROM porte_identities i WHERE i.user_id = $1 AND strpos(o.payload::text, '"' || i.subject || '"') > 0)`
)

// Eraser deletes what an erased person owns, as the user store's EraseData:
// it runs in the erase transaction, before the users row goes. Deleting that
// row removes the personal workspace with its projects, tasks and files, the
// memberships, notifications, preferences, invitations the person sent,
// sessions and API tokens. What they wrote in shared workspaces stays:
// comments, activity, time entries, assets and share links lose their
// author. See docs/erasure.md.
type Eraser struct {
	uploadDir string
	assets    AssetRemover
}

// NewEraser returns the eraser over the upload directory and the asset store,
// on local disk or S3.
func NewEraser(uploadDir string, assets AssetRemover) *Eraser {
	return &Eraser{uploadDir: uploadDir, assets: assets}
}

// EraseData removes the person's stored objects first, then the invitations
// sent to their address, their pending MCP actions and every bus event that
// names them, sent or still queued, and drops their GitHub id from the pull
// requests they authored. A queued event dropped here never reaches Antenne;
// the next reconcile re-sends the current state of shared projects and tasks. A storage failure rolls the transaction back, so
// Registre's retry finds the account and deletes the objects again; deleting
// a missing object succeeds.
func (e *Eraser) EraseData(ctx context.Context, tx *sql.Tx, userID int64) error {
	keys, err := queryColumn[string](ctx, tx, selectPersonalAssets, userID)
	if err != nil {
		return err
	}
	personal, err := queryColumn[int64](ctx, tx, selectPersonalWorkspace, userID)
	if err != nil {
		return err
	}
	if err := e.assets.Delete(ctx, keys); err != nil {
		return err
	}
	for _, id := range personal {
		if err := storage.DeleteWorkspaceLogo(e.uploadDir, id); err != nil {
			slog.Warn("failed to remove an erased person's workspace logo", slog.Any("error", err))
		}
	}
	for _, statement := range []string{deleteInvitationsToThem, forgetPullAuthor, deleteEventsNamingThem} {
		if _, err := tx.ExecContext(ctx, statement, userID); err != nil {
			return err
		}
	}
	return mcpkit.ErasePending(ctx, tx, strconv.FormatInt(userID, 10))
}

func queryColumn[T any](ctx context.Context, tx *sql.Tx, query string, args ...any) ([]T, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []T
	for rows.Next() {
		var value T
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

// deletionState keeps the receiver's cursor and handled subjects in one row.
type deletionState struct {
	orm *gorm.DB
}

// Load returns the saved state, or nil before the first save.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var rows []struct{ State []byte }
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0].State, nil
}

// Save replaces the saved state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	return s.orm.WithContext(ctx).Exec(
		`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())
		 ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, state).Error
}
