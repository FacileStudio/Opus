package workspace

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// personalName names the workspace every user owns from their first visit.
const personalName = "Personal"

// ensurePersonal gives the user their personal workspace when they have none
// yet, which is how a new account starts with one without creating anything.
// It is private to its owner: never imported into Registre, never linked to a
// shared space, never shared or deleted.
func (service *Service) ensurePersonal(ctx context.Context, userID int64) error {
	return service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&schemas.Workspace{}).Where("personal_owner_id = ?", userID).Count(&count).Error; err != nil || count > 0 {
			return err
		}
		if err := tx.Model(&schemas.User{}).Where("id = ?", userID).Count(&count).Error; err != nil || count == 0 {
			return err
		}
		slug, err := uniqueSlug(tx, "personal")
		if err != nil {
			return err
		}
		owner := userID
		workspace := schemas.Workspace{Name: personalName, Slug: slug, CreatedAt: time.Now(), PersonalOwnerID: &owner}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&workspace)
		if created.Error != nil || created.RowsAffected == 0 {
			return created.Error
		}
		return tx.Create(&schemas.WorkspaceMember{WorkspaceID: workspace.ID, UserID: userID, Role: "owner", JoinedAt: time.Now()}).Error
	})
}

// refusePersonal answers a conflict for an action that would share or delete
// a personal workspace.
func (service *Service) refusePersonal(ctx context.Context, workspaceID int64) error {
	var workspace schemas.Workspace
	err := service.orm.WithContext(ctx).Select("personal_owner_id").First(&workspace, "id = ?", workspaceID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return errors.Internal("failed to load the workspace", err)
	}
	if workspace.PersonalOwnerID != nil {
		return errors.Conflict("Your personal workspace stays private to you and cannot be deleted")
	}
	return nil
}

// requireShareable refuses inviting into or deleting a workspace that is
// either owned by Registre or someone's personal workspace.
func (service *Service) requireShareable(ctx context.Context, workspaceID int64) error {
	if err := service.requireLocal(ctx, workspaceID); err != nil {
		return err
	}
	return service.refusePersonal(ctx, workspaceID)
}
