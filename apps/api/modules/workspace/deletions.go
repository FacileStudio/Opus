package workspace

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	orm       *gorm.DB
	uploadDir string
	receiver  *directory.Deletions
}

// NewDeletions builds the receiver over Opus's tables. erase is porte's own
// rows eraser, pg.Store.EraseIdentity.
func NewDeletions(
	service *Service, client directory.Client, erase func(context.Context, string, string) error,
	logger *slog.Logger,
) *Deletions {
	d := &Deletions{orm: service.orm, uploadDir: service.uploadDir}
	d.receiver = &directory.Deletions{
		Client:          client,
		OnPersonDeleted: d.ErasePerson,
		Erase:           erase,
		State:           deletionState{orm: service.orm},
		Logger:          logger,
	}
	return d
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
	selectErasedAccount = `SELECT u.id, u.email FROM users u
		JOIN porte_identities i ON i.user_id = u.id WHERE i.provider = ? AND i.subject = ?`
	selectPersonalAssets = `SELECT a.object_key FROM asset a
		JOIN workspaces w ON w.id = a.workspace_id WHERE w.personal_owner_id = ?`
	deleteInvitationsToThem = `DELETE FROM invitations WHERE lower(email) = lower(?)`
	deleteAccount           = `DELETE FROM users WHERE id = ?`
)

// ErasePerson deletes the account behind a Registre identity and what it owns.
// Deleting the users row removes the personal workspace with its projects,
// tasks and files, the memberships, notifications, preferences, invitations
// the person sent, sessions and API tokens. What they wrote in shared
// workspaces stays: comments, activity, time entries, assets and share links
// lose their author, which the client shows as a deleted user. Invitations
// sent to their address go too. A subject that is already gone is a no-op, so
// a replayed notice changes nothing.
func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	var account struct {
		ID    int64
		Email string
	}
	if err := d.orm.WithContext(ctx).Raw(selectErasedAccount, provider, subject).Scan(&account).Error; err != nil {
		return err
	}
	if account.ID == 0 {
		return nil
	}
	var keys []string
	err := d.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(selectPersonalAssets, account.ID).Scan(&keys).Error; err != nil {
			return err
		}
		if err := tx.Exec(deleteInvitationsToThem, account.Email).Error; err != nil {
			return err
		}
		return tx.Exec(deleteAccount, account.ID).Error
	})
	if err == nil {
		d.removeAssets(keys)
	}
	return err
}

func (d *Deletions) removeAssets(keys []string) {
	for _, key := range keys {
		path := filepath.Join(d.uploadDir, "assets", filepath.Clean("/"+key))
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("failed to remove an erased person's asset", slog.Any("error", err))
		}
	}
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
