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
	case existing == nil && action == "created":
		return s.createProject(tx, evt)
	case existing == nil, action == "created":
		return outcome{}, nil
	default:
		return outcome{}, updateProject(tx, existing, evt)
	}
}

func (s *Syncer) createProject(tx *gorm.DB, evt inbound) (outcome, error) {
	if s.workspaceID == 0 || evt.Name == "" {
		s.logger.Warn("pool sync: project not created, no target workspace or no name",
			slog.String("facile_id", evt.facileID))
		return outcome{skipped: true}, nil
	}
	var count int64
	if err := tx.Model(&schemas.Workspace{}).Where("id = ?", s.workspaceID).Count(&count).Error; err != nil {
		return outcome{}, err
	}
	if count == 0 {
		s.logger.Warn("pool sync: target workspace not found", slog.Int64("workspace_id", s.workspaceID))
		return outcome{skipped: true}, nil
	}
	slug, err := uniqueSlug(tx, s.workspaceID, project.Slugify(evt.Name))
	if err != nil {
		return outcome{}, err
	}
	facileID := evt.facileID
	record := schemas.Project{
		WorkspaceID: s.workspaceID,
		FacileRef:   schemas.FacileRef{FacileID: &facileID},
		Slug:        slug,
		Icon:        normalizeIcon(evt.Icon),
		Name:        evt.Name,
		CreatedAt:   time.Now(),
	}
	if evt.Description != nil {
		record.Description = *evt.Description
	}
	if err := tx.Create(&record).Error; err != nil {
		return outcome{}, err
	}
	return outcome{}, project.CreateDefaultColumns(tx, record.ID, record.CreatedAt)
}

func updateProject(tx *gorm.DB, existing *schemas.Project, evt inbound) error {
	updates := map[string]any{}
	if evt.Name != "" {
		updates["name"] = evt.Name
	}
	if evt.Description != nil {
		updates["description"] = *evt.Description
	}
	if evt.Icon != nil {
		updates["icon"] = normalizeIcon(evt.Icon)
	}
	if len(updates) == 0 {
		return nil
	}
	return tx.Model(&schemas.Project{}).Where("id = ?", existing.ID).Updates(updates).Error
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
