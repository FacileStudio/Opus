package poolsync

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/modules/project"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func findProject(tx *gorm.DB, facileID string) (*schemas.Project, error) {
	var found []schemas.Project
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("facile_id = ?", facileID).Limit(1).Find(&found).Error; err != nil || len(found) == 0 {
		return nil, err
	}
	return &found[0], nil
}

func (s *Syncer) applyProject(tx *gorm.DB, action string, evt inbound) (outcome, error) {
	existing, err := findProject(tx, evt.facileID)
	if err != nil {
		return outcome{}, err
	}
	switch {
	case action == "deleted":
		return outcome{}, deleteProject(tx, existing)
	case existing != nil:
		return s.updateProject(tx, existing, evt)
	case action == "created":
		return s.createProject(tx, evt)
	default:
		return outcome{}, nil
	}
}

func (s *Syncer) unplaced(evt inbound, reason string) (outcome, error) {
	s.logger.Warn("pool sync: event skipped, "+reason, slog.String("facile_id", evt.facileID))
	return outcome{skipped: true}, nil
}

func (s *Syncer) createProject(tx *gorm.DB, evt inbound) (outcome, error) {
	if evt.Name == "" {
		return s.unplaced(evt, "project has no name")
	}
	workspaceID, err := s.resolveWorkspace(tx, evt.Space)
	if err != nil {
		return outcome{}, err
	}
	if workspaceID == 0 {
		return s.unplaced(evt, "project space does not resolve to a workspace")
	}
	var twins []schemas.Project
	if err := tx.Where("workspace_id = ? AND lower(name) = lower(?) AND facile_id <> ? AND synced_at IS NULL",
		workspaceID, evt.Name, evt.facileID).Order("id").Limit(1).Find(&twins).Error; err != nil {
		return outcome{}, err
	}
	if len(twins) > 0 {
		return s.mergeProject(tx, &twins[0], evt)
	}
	slug, err := uniqueSlug(tx, workspaceID, project.Slugify(evt.Name))
	if err != nil {
		return outcome{}, err
	}
	facileID, now := evt.facileID, time.Now()
	record := schemas.Project{
		WorkspaceID: workspaceID,
		FacileRef:   schemas.FacileRef{FacileID: &facileID},
		Slug:        slug,
		Icon:        normalizeIcon(evt.Icon),
		Name:        evt.Name,
		CreatedAt:   now,
		UpdatedAt:   evt.updated,
		SyncedAt:    &now,
	}
	if evt.Description != nil {
		record.Description = *evt.Description
	}
	if err := tx.Create(&record).Error; err != nil {
		return outcome{}, err
	}
	return outcome{}, project.CreateDefaultColumns(tx, record.ID, record.CreatedAt)
}

// mergeProject links an unlinked local twin to the incoming project. Both
// sides keep the smaller facile_id, so the twin adopts the incoming one only
// when it sorts first; otherwise the event is dropped and Sablier adopts ours.
func (s *Syncer) mergeProject(tx *gorm.DB, twin *schemas.Project, evt inbound) (outcome, error) {
	if twin.FacileID != nil && evt.facileID >= *twin.FacileID {
		return outcome{}, nil
	}
	if err := tx.Model(&schemas.Project{}).Where("id = ?", twin.ID).
		Updates(map[string]any{"facile_id": evt.facileID, "updated_at": twin.UpdatedAt, "synced_at": time.Now()}).Error; err != nil {
		return outcome{}, err
	}
	twin.FacileID = &evt.facileID
	out, err := s.updateProject(tx, twin, evt)
	if err != nil || out.skipped {
		return out, err
	}
	return outcome{adopted: true, projectID: twin.ID}, nil
}

func (s *Syncer) updateProject(tx *gorm.DB, existing *schemas.Project, evt inbound) (outcome, error) {
	if evt.updated.Before(existing.UpdatedAt) {
		return outcome{}, nil
	}
	updates := map[string]any{"updated_at": evt.updated, "synced_at": time.Now()}
	if evt.Name != "" {
		updates["name"] = evt.Name
	}
	if evt.Description != nil {
		updates["description"] = *evt.Description
	}
	if evt.Icon != nil {
		updates["icon"] = normalizeIcon(evt.Icon)
	}
	if evt.Space != nil {
		target, err := s.resolveWorkspace(tx, evt.Space)
		if err != nil {
			return outcome{}, err
		}
		if target == 0 {
			return s.unplaced(evt, "project space does not resolve to a workspace")
		}
		if err := moveProject(tx, existing, target, updates); err != nil {
			return outcome{}, err
		}
	}
	return outcome{}, tx.Model(&schemas.Project{}).Where("id = ?", existing.ID).Updates(updates).Error
}

// moveProject moves a project, and its tasks with it, to another workspace,
// renaming its slug when the target already uses it.
func moveProject(tx *gorm.DB, existing *schemas.Project, target int64, updates map[string]any) error {
	if target == existing.WorkspaceID {
		return nil
	}
	slug, err := uniqueSlug(tx, target, existing.Slug)
	if err != nil {
		return err
	}
	updates["workspace_id"] = target
	updates["slug"] = slug
	return tx.Model(&schemas.Task{}).Where("project_id = ?", existing.ID).Update("workspace_id", target).Error
}

func deleteProject(tx *gorm.DB, existing *schemas.Project) error {
	if existing == nil {
		return nil
	}
	for _, model := range []any{&schemas.Task{}, &schemas.Column{}} {
		if err := tx.Where("project_id = ?", existing.ID).Delete(model).Error; err != nil {
			return err
		}
	}
	return tx.Where("id = ?", existing.ID).Delete(&schemas.Project{}).Error
}

func uniqueSlug(tx *gorm.DB, workspaceID int64, base string) (string, error) {
	candidate := base
	for suffix := 2; ; suffix++ {
		var count int64
		err := tx.Model(&schemas.Project{}).Where("workspace_id = ? AND slug = ?", workspaceID, candidate).Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, suffix)
	}
}
