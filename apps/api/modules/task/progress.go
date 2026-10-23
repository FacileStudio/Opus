package task

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// SetPublisher enables live broadcasts of task changes.
func (service *Service) SetPublisher(publisher realtime.Publisher) {
	service.publisher = publisher
}

func (service *Service) publish(ctx context.Context, msgType string, projectID, taskID int64) {
	realtime.Notify(ctx, service.publisher, msgType, projectID, taskID)
}

func (service *Service) projectIDsOf(ctx context.Context, taskIDs []wireid.ID) []int64 {
	ids := make([]int64, len(taskIDs))
	for i, id := range taskIDs {
		ids[i] = int64(id)
	}
	var projectIDs []int64
	service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id IN ?", ids).Distinct().Pluck("project_id", &projectIDs)
	return projectIDs
}

type subtaskProgress struct {
	total int
	done  int
}

func fetchSubtaskProgress(ctx context.Context, db *gorm.DB, taskIDs []int64) (map[int64]subtaskProgress, error) {
	var rows []struct {
		SourceTaskID int64
		Total        int
		Done         int
	}
	err := db.WithContext(ctx).
		Table("task_relations AS r").
		Select("r.source_task_id AS source_task_id, COUNT(*) AS total, SUM(CASE WHEN t.status IN ('done', 'archived') THEN 1 ELSE 0 END) AS done").
		Joins("JOIN tasks AS t ON t.id = r.target_task_id").
		Where("r.relation_type = ? AND r.source_task_id IN ?", "subtask", taskIDs).
		Group("r.source_task_id").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to load subtask progress", err)
	}
	out := make(map[int64]subtaskProgress, len(rows))
	for _, row := range rows {
		out[row.SourceTaskID] = subtaskProgress{total: row.Total, done: row.Done}
	}
	return out, nil
}
