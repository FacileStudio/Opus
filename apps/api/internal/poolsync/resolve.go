package poolsync

import (
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

func firstID(query *gorm.DB) (int64, error) {
	var ids []int64
	if err := query.Limit(1).Pluck("id", &ids).Error; err != nil || len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}

// resolveWorkspace finds the workspace a project's space maps to: the one
// linked to its Registre space, or its owner's oldest personal workspace. A
// payload with no space, from an older Sablier, falls back to the configured
// sync workspace. It answers 0 when nothing matches.
func (s *Syncer) resolveWorkspace(tx *gorm.DB, space *spaceBody) (int64, error) {
	workspaces := tx.Model(&schemas.Workspace{})
	switch {
	case space == nil && s.workspaceID == 0:
		return 0, nil
	case space == nil:
		return firstID(workspaces.Where("id = ?", s.workspaceID))
	case space.RegistreSpaceID != "":
		return firstID(workspaces.Where("registre_space_id = ?", space.RegistreSpaceID))
	case !space.Personal:
		return 0, nil
	}
	owner, err := s.resolveUser(tx, space.OwnerSubject, space.OwnerEmail)
	if err != nil || owner == 0 {
		return 0, err
	}
	return firstID(workspaces.Where("personal_owner_id = ?", owner).Order("id"))
}

// resolveUser finds a user by porte subject under the issuer, then by email
// ignoring case. It answers 0 when neither matches.
func (s *Syncer) resolveUser(tx *gorm.DB, subject, email string) (int64, error) {
	subject, email = strings.TrimSpace(subject), strings.TrimSpace(email)
	if subject != "" && s.issuer != "" {
		var ids []int64
		err := tx.Table("porte_identities").Where("provider = ? AND subject = ?", s.issuer, subject).
			Limit(1).Pluck("user_id", &ids).Error
		if err != nil {
			return 0, err
		}
		if len(ids) > 0 {
			return ids[0], nil
		}
	}
	if email == "" {
		return 0, nil
	}
	return firstID(tx.Model(&schemas.User{}).Where("lower(email) = lower(?)", email).Order("id"))
}
