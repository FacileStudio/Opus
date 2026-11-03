package events

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

const reconcileBatch = 500

// Reconcile is the contract's backfill: it assigns missing facile_ids, then
// queues project.created for every active project Sablier can place, then
// task.created for every task of those projects. Sablier treats a created for
// a known facile_id as an update and merges name twins, so running it again
// is safe.
func (p *Publisher) Reconcile(ctx context.Context) error {
	if !p.active() || p.db == nil {
		return nil
	}
	for _, model := range []any{&schemas.Project{}, &schemas.Task{}} {
		if err := p.backfillFacileIDs(ctx, model); err != nil {
			return err
		}
	}
	var projects []schemas.Project
	if err := p.db.WithContext(ctx).Where("archived_at IS NULL").Order("id").Find(&projects).Error; err != nil {
		return err
	}
	facileIDs := map[int64]string{}
	for _, project := range projects {
		if _, ok := p.SpaceOf(ctx, project.WorkspaceID); !ok {
			continue
		}
		facileIDs[project.ID] = facileIDOf(project.FacileID)
		if err := p.emitProject(ctx, ChannelProjectCreated, NewProjectPayload(project, ""), false); err != nil {
			return err
		}
	}
	queued, err := p.reconcileTasks(ctx, facileIDs)
	p.logger.Info("antenne: reconcile queued", slog.Int("projects", len(facileIDs)), slog.Int("tasks", queued))
	return err
}

// ReemitTasks queues task.created for every task of a project, so Sablier
// learns them under the project's current facile_id after a twin merge.
func (p *Publisher) ReemitTasks(ctx context.Context, projectID int64) {
	if !p.active() || p.db == nil {
		return
	}
	var facileIDs []string
	p.db.WithContext(ctx).Model(&schemas.Project{}).Where("id = ?", projectID).Limit(1).Pluck("facile_id", &facileIDs)
	if len(facileIDs) == 0 {
		return
	}
	if _, err := p.reconcileTasks(ctx, map[int64]string{projectID: facileIDs[0]}); err != nil {
		p.logger.Warn("antenne: task re-emit failed", slog.Int64("project_id", projectID), slog.Any("error", err))
	}
}

func (p *Publisher) reconcileTasks(ctx context.Context, facileIDs map[int64]string) (int, error) {
	if len(facileIDs) == 0 {
		return 0, nil
	}
	projectIDs := make([]int64, 0, len(facileIDs))
	for id := range facileIDs {
		projectIDs = append(projectIDs, id)
	}
	queued := 0
	var batch []schemas.Task
	err := p.db.WithContext(ctx).Where("project_id IN ?", projectIDs).
		FindInBatches(&batch, reconcileBatch, func(tx *gorm.DB, _ int) error {
			emails := TaskActorEmails(ctx, p.db, batch)
			for _, task := range batch {
				payload := NewTaskPayload(task, facileIDs[task.ProjectID], emails[task.ID])
				if err := p.emitTask(ctx, ChannelTaskCreated, payload, false); err != nil {
					return err
				}
				queued++
			}
			return nil
		}).Error
	return queued, err
}

func (p *Publisher) backfillFacileIDs(ctx context.Context, model any) error {
	var ids []int64
	if err := p.db.WithContext(ctx).Model(model).Where("facile_id IS NULL OR facile_id = ''").Pluck("id", &ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		p.ensureFacileID(ctx, model, id, "")
	}
	return nil
}
