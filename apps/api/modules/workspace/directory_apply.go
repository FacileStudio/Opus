package workspace

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Apply makes the workspaces match a Registre snapshot in one transaction. It
// is idempotent: applying the same snapshot twice changes nothing.
//
// Each space links to the workspace already carrying its id, else to an
// unlinked workspace with the same name ignoring case, else to a new one. Each
// member resolves by identity, then by email, and is otherwise pre-provisioned
// as a user with an identity row, so their first sign-in lands on that user.
// Memberships of users with a Registre identity who are absent from the space
// are removed; members without one, unlinked workspaces and linked workspaces
// missing from the snapshot are left alone. Pending invitations to linked
// workspaces are canceled, since adding people is Registre's job now.
func (d *Directory) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	if d.onSnapshot != nil {
		d.onSnapshot(snapshot)
	}
	listed := []string{}
	err := d.service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, space := range snapshot.Spaces {
			if err := d.applySpace(tx, space); err != nil {
				return err
			}
			listed = append(listed, space.ID)
		}
		if err := tx.Model(&schemas.Invitation{}).
			Where("status = ? AND workspace_id IN (?)", "pending",
				tx.Model(&schemas.Workspace{}).Select("id").Where("registre_space_id IS NOT NULL")).
			Update("status", "canceled").Error; err != nil {
			return errors.Internal("failed to cancel the invitations of linked workspaces", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	d.logMissing(ctx, listed)
	d.notifySynced(ctx)
	return nil
}

// applySpace links one space and makes its memberships match.
func (d *Directory) applySpace(tx *gorm.DB, space directory.Space) error {
	workspaceID, err := d.linkSpace(tx, space)
	if err != nil {
		return err
	}
	wanted := map[int64]string{}
	for _, member := range space.Members {
		userID, err := d.resolveMember(tx, member)
		if err != nil {
			return err
		}
		if userID != 0 {
			wanted[userID] = registreRole(member.Role)
		}
	}
	return d.syncMembers(tx, workspaceID, wanted)
}

// linkSpace finds or creates the workspace a space mirrors and gives it
// Registre's name.
func (d *Directory) linkSpace(tx *gorm.DB, space directory.Space) (int64, error) {
	var workspace schemas.Workspace
	found := tx.Where("registre_space_id = ?", space.ID).Limit(1).Find(&workspace)
	if found.Error == nil && found.RowsAffected == 0 {
		found = tx.Where("registre_space_id IS NULL AND personal_owner_id IS NULL AND lower(name) = lower(?)", space.Name).
			Order("id").Limit(1).Find(&workspace)
	}
	if found.Error != nil {
		return 0, errors.Internal("failed to look up the workspace of a space", found.Error)
	}
	if found.RowsAffected == 0 {
		return d.createLinked(tx, space)
	}
	if workspace.Name == space.Name && workspace.RegistreSpaceID != nil {
		return workspace.ID, nil
	}
	if err := tx.Model(&schemas.Workspace{}).Where("id = ?", workspace.ID).
		Updates(map[string]any{"name": space.Name, "registre_space_id": space.ID}).Error; err != nil {
		return 0, errors.Internal("failed to link the workspace", err)
	}
	return workspace.ID, nil
}

// createLinked creates the workspace for a space Opus has never held.
func (d *Directory) createLinked(tx *gorm.DB, space directory.Space) (int64, error) {
	slug, err := uniqueSlug(tx, slugify(space.Name))
	if err != nil {
		return 0, err
	}
	spaceID := space.ID
	workspace := schemas.Workspace{Name: space.Name, Slug: slug, CreatedAt: time.Now(), RegistreSpaceID: &spaceID}
	if err := tx.Create(&workspace).Error; err != nil {
		return 0, errors.Internal("failed to create the workspace of a space", err)
	}
	return workspace.ID, nil
}

// resolveMember finds the user a Registre member is, by identity and then by
// email, and pre-provisions one otherwise. It answers 0 for a member Registre
// sent without a subject, which cannot be matched safely.
func (d *Directory) resolveMember(tx *gorm.DB, member directory.Member) (int64, error) {
	if member.Subject == "" {
		return 0, nil
	}
	var userID int64
	if err := tx.Table("porte_identities").Select("user_id").
		Where("provider = ? AND subject = ?", d.issuer, member.Subject).
		Limit(1).Scan(&userID).Error; err != nil {
		return 0, errors.Internal("failed to resolve a member identity", err)
	}
	if userID != 0 {
		return userID, nil
	}
	email := strings.ToLower(strings.TrimSpace(member.Email))
	if email == "" {
		return 0, nil
	}
	if err := tx.Model(&schemas.User{}).Select("id").Where("email = ?", email).
		Limit(1).Scan(&userID).Error; err != nil {
		return 0, errors.Internal("failed to resolve a member email", err)
	}
	if userID == 0 {
		user := schemas.User{Email: email, EmailVerified: true, Name: member.Name, Image: httpsOnly(member.Picture)}
		if err := tx.Create(&user).Error; err != nil {
			return 0, errors.Internal("failed to pre-provision a member", err)
		}
		userID = user.ID
	}
	if err := tx.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)",
		userID, d.issuer, member.Subject).Error; err != nil {
		return 0, errors.Internal("failed to record a member identity", err)
	}
	return userID, nil
}

// syncMembers adds and re-ranks the wanted members of a workspace, and
// removes members with a Registre identity who are no longer wanted.
func (d *Directory) syncMembers(tx *gorm.DB, workspaceID int64, wanted map[int64]string) error {
	current := []schemas.WorkspaceMember{}
	if err := tx.Where("workspace_id = ?", workspaceID).Find(&current).Error; err != nil {
		return errors.Internal("failed to read the workspace members", err)
	}
	held := map[int64]string{}
	for _, member := range current {
		held[member.UserID] = member.Role
	}
	for userID, role := range wanted {
		if err := setMember(tx, workspaceID, userID, role, held); err != nil {
			return err
		}
	}
	gone := []int64{}
	for userID := range held {
		if _, keep := wanted[userID]; !keep {
			gone = append(gone, userID)
		}
	}
	if len(gone) == 0 {
		return nil
	}
	if err := tx.Where("workspace_id = ? AND user_id IN ? AND user_id IN (?)", workspaceID, gone,
		tx.Table("porte_identities").Select("user_id").Where("provider = ?", d.issuer)).
		Delete(&schemas.WorkspaceMember{}).Error; err != nil {
		return errors.Internal("failed to remove departed members", err)
	}
	return nil
}

// setMember inserts a membership or corrects its role, and does nothing when
// it already matches.
func setMember(tx *gorm.DB, workspaceID, userID int64, role string, held map[int64]string) error {
	existing, ok := held[userID]
	switch {
	case !ok:
		member := schemas.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: role, JoinedAt: time.Now()}
		if err := tx.Create(&member).Error; err != nil {
			return errors.Internal("failed to add a member", err)
		}
	case existing != role:
		if err := tx.Model(&schemas.WorkspaceMember{}).
			Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
			Update("role", role).Error; err != nil {
			return errors.Internal("failed to change a member role", err)
		}
	}
	return nil
}

// logMissing reports linked workspaces Registre no longer lists. Registre
// cannot delete spaces yet, so they are kept as they are.
func (d *Directory) logMissing(ctx context.Context, listed []string) {
	missing := []string{}
	query := d.service.orm.WithContext(ctx).Model(&schemas.Workspace{}).Where("registre_space_id IS NOT NULL")
	if len(listed) > 0 {
		query = query.Where("registre_space_id NOT IN ?", listed)
	}
	if err := query.Pluck("registre_space_id", &missing).Error; err != nil || len(missing) == 0 {
		return
	}
	d.logger.Info("linked workspaces missing from the registre snapshot, kept", slog.Any("spaces", missing))
}

// httpsOnly keeps a picture URL the browser can load on an https page.
func httpsOnly(url string) string {
	if strings.HasPrefix(strings.ToLower(url), "https://") {
		return url
	}
	return ""
}
