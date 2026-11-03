package workspace

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/errors"
)

const (
	firstImportRetry = 5 * time.Second
	maxImportRetry   = 5 * time.Minute
)

// Directory keeps Opus workspaces equal to Registre's shared spaces. Once the
// one-time import has reached Registre, Registre owns every workspace's name
// and members, and Opus only applies the snapshots it serves.
type Directory struct {
	service    *Service
	issuer     string
	client     directory.Client
	syncer     *directory.Syncer
	logger     *slog.Logger
	retryDelay time.Duration
	synced     sync.Once
	onSynced   func(context.Context)
}

// OnFirstSync runs fn once, in the background, after the first Registre
// snapshot has been applied, when every shared workspace carries its
// Registre space id. Call it before Run.
func (d *Directory) OnFirstSync(fn func(context.Context)) {
	d.onSynced = fn
}

func (d *Directory) notifySynced(ctx context.Context) {
	if d.onSynced == nil {
		return
	}
	d.synced.Do(func() { go d.onSynced(context.WithoutCancel(ctx)) })
}

// NewDirectory wires the Registre directory sync for the workspace service.
// The client carries the issuer and the OIDC client credentials Opus already
// signs users in with. From here on, edits to a linked workspace are refused
// with a pointer to its Registre page.
func NewDirectory(service *Service, client directory.Client, logger *slog.Logger) *Directory {
	d := &Directory{
		service:    service,
		issuer:     client.Issuer,
		client:     client,
		logger:     logger,
		retryDelay: firstImportRetry,
	}
	d.syncer = &directory.Syncer{Client: client, Apply: d.Apply, Logger: logger}
	service.registreSpaces = strings.TrimSuffix(client.Issuer, "/") + "/spaces"
	return d
}

// Handler answers Registre's nudge on directory.SyncPath.
func (d *Directory) Handler() http.Handler {
	return d.syncer.Handler()
}

// Run imports the existing workspaces once, retrying with backoff while
// Registre is unreachable or does not serve the directory yet, and only then
// starts the periodic sync. Syncing first could remove members before their
// workspace ever reached Registre. Run returns when ctx ends.
func (d *Directory) Run(ctx context.Context) {
	if d.service.hasLinkedWorkspace(ctx) {
		d.service.managed.Store(true)
	}
	if !d.importUntilDone(ctx) {
		return
	}
	d.service.managed.Store(true)
	d.syncer.Run(ctx)
}

// importUntilDone calls Import until it succeeds, doubling the wait up to
// maxImportRetry, and reports false when ctx ended first.
func (d *Directory) importUntilDone(ctx context.Context) bool {
	delay := d.retryDelay
	for {
		err := d.importOnce(ctx)
		if err == nil {
			return true
		}
		d.logger.Warn("registre import failed, retrying", slog.Any("error", err), slog.Duration("retry_in", delay))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
		delay = min(delay*2, maxImportRetry)
	}
}

// importOnce reads the current workspaces and hands them to Registre. They are
// read again on every attempt, so changes made while Registre was unreachable
// are not lost.
func (d *Directory) importOnce(ctx context.Context) error {
	spaces, err := d.service.importSpaces(ctx, d.issuer)
	if err != nil {
		return err
	}
	_, err = d.client.Import(ctx, spaces)
	return err
}

type importRow struct {
	WorkspaceID   int64
	Name          string
	Role          string
	Email         string
	EmailVerified bool
	Subject       *string
}

// importSpaces lists every workspace with the members Registre can identify:
// by their subject under issuer, or else by a verified email. A workspace
// with no such member is left out and stays a local workspace.
func (service *Service) importSpaces(ctx context.Context, issuer string) ([]directory.ImportSpace, error) {
	rows := []importRow{}
	err := service.orm.WithContext(ctx).
		Table("workspaces").
		Select("workspaces.id AS workspace_id, workspaces.name, workspace_members.role, users.email, "+
			"users.email_verified, porte_identities.subject").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Joins("JOIN users ON users.id = workspace_members.user_id").
		Joins("LEFT JOIN porte_identities ON porte_identities.user_id = users.id AND porte_identities.provider = ?", issuer).
		Where("workspaces.registre_space_id IS NULL AND workspaces.personal_owner_id IS NULL").
		Order("workspaces.id").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to read the workspaces to import", err)
	}
	spaces := []directory.ImportSpace{}
	index := map[int64]int{}
	for _, row := range rows {
		member, ok := importMember(row)
		if !ok {
			continue
		}
		position, seen := index[row.WorkspaceID]
		if !seen {
			position = len(spaces)
			index[row.WorkspaceID] = position
			spaces = append(spaces, directory.ImportSpace{Name: row.Name})
		}
		spaces[position].Members = append(spaces[position].Members, member)
	}
	return spaces, nil
}

// importMember is the Registre-side identity of one member, or false when
// Registre has no safe way to recognise them.
func importMember(row importRow) (directory.ImportMember, bool) {
	role := registreRole(row.Role)
	if row.Subject != nil && *row.Subject != "" {
		return directory.ImportMember{Subject: *row.Subject, Role: role}, true
	}
	if row.EmailVerified && row.Email != "" {
		return directory.ImportMember{Email: row.Email, Role: role}, true
	}
	return directory.ImportMember{}, false
}

// registreRole maps an Opus role onto Registre's three. Opus uses the same
// vocabulary, so anything unexpected falls back to member.
func registreRole(role string) string {
	switch role {
	case directory.RoleOwner, directory.RoleAdmin:
		return role
	default:
		return directory.RoleMember
	}
}

// SpacesURL is Registre's space management page once Registre owns
// workspaces, and empty while Opus still manages them itself.
func (service *Service) SpacesURL() string {
	if !service.managed.Load() {
		return ""
	}
	return service.registreSpaces
}

// hasLinkedWorkspace reports whether an earlier run already linked a
// workspace to Registre, which means Registre owned workspaces before.
func (service *Service) hasLinkedWorkspace(ctx context.Context) bool {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("registre_space_id IS NOT NULL").Count(&count).Error
	return err == nil && count > 0
}

// requireLocal refuses to edit a workspace Registre owns, pointing the caller
// at the Registre page where the change belongs.
func (service *Service) requireLocal(ctx context.Context, workspaceID int64) error {
	var linked *string
	err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("id = ?", workspaceID).Select("registre_space_id").Scan(&linked).Error
	if err != nil {
		return errors.Internal("failed to load the workspace", err)
	}
	if linked == nil || *linked == "" {
		return nil
	}
	return errors.Conflict("This workspace is managed in Registre: " + service.spaceURL(*linked))
}

// spaceURL is the Registre page of one linked space.
func (service *Service) spaceURL(spaceID string) string {
	return service.registreSpaces + "/" + spaceID
}
