package task

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Search sort orders.
const (
	SortUpdated  = "updated"
	SortDue      = "due"
	SortPriority = "priority"
	SortCreated  = "created"
	SortNumber   = "number"
)

// SearchQuery filters tasks across the workspaces a person belongs to. Every
// field is optional; the zero value lists every open task the person can
// see, most recently updated first.
type SearchQuery struct {
	WorkspaceID     int64
	ProjectID       int64
	Text            string
	Statuses        []string
	Priority        string
	AssigneeID      *int64
	Unassigned      bool
	Label           string
	DueBefore       *time.Time
	DueAfter        *time.Time
	IncludeDone     bool
	IncludeArchived bool
	Sort            string
	Offset          int
	Limit           int
}

// Search returns the tasks matching q in the workspaces userID is a member
// of, with their total count before paging. Naming a workspace the person is
// not a member of is refused, as every task route refuses it.
func (service *Service) Search(ctx context.Context, userID int64, q SearchQuery) ([]TaskRecord, int64, error) {
	if q.WorkspaceID != 0 {
		if err := service.verifyWorkspaceMembership(ctx, userID, q.WorkspaceID); err != nil {
			return nil, 0, err
		}
	}
	if q.ProjectID != 0 {
		if _, err := service.verifyProjectMembership(ctx, userID, q.ProjectID); err != nil {
			return nil, 0, err
		}
	}
	query := searchFilters(service.db.WithContext(ctx).Model(&schemas.Task{}), userID, q)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Internal("failed to count tasks", err)
	}
	var tasks []schemas.Task
	err := query.Order(searchOrder(q.Sort)).Order("tasks.id desc").Offset(q.Offset).Limit(q.Limit).Find(&tasks).Error
	if err != nil {
		return nil, 0, errors.Internal("failed to search tasks", err)
	}
	records, err := service.loadTaskRecords(ctx, tasks)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// searchFilters scopes query to userID's workspaces and applies q.
func searchFilters(query *gorm.DB, userID int64, q SearchQuery) *gorm.DB {
	query = query.Where("tasks.workspace_id IN (SELECT workspace_id FROM workspace_members WHERE user_id = ?)", userID)
	if q.WorkspaceID != 0 {
		query = query.Where("tasks.workspace_id = ?", q.WorkspaceID)
	}
	if q.ProjectID != 0 {
		query = query.Where("tasks.project_id = ?", q.ProjectID)
	}
	for _, word := range strings.Fields(strings.ToLower(q.Text)) {
		like := "%" + escapeLike(word) + "%"
		query = query.Where("(LOWER(tasks.title) LIKE ? ESCAPE '\\' OR LOWER(tasks.description) LIKE ? ESCAPE '\\')", like, like)
	}
	return lifecycleFilters(fieldFilters(query, q), q)
}

// fieldFilters applies the per-field filters of q.
func fieldFilters(query *gorm.DB, q SearchQuery) *gorm.DB {
	if len(q.Statuses) > 0 {
		query = query.Where("tasks.status IN ?", q.Statuses)
	}
	if q.Priority != "" {
		query = query.Where("tasks.priority = ?", q.Priority)
	}
	if q.Unassigned {
		query = query.Where("tasks.actor_id IS NULL")
	} else if q.AssigneeID != nil {
		query = query.Where("tasks.actor_id = ?", *q.AssigneeID)
	}
	if q.Label != "" {
		query = query.Where("EXISTS (SELECT 1 FROM labels l WHERE l.task_id = tasks.id AND LOWER(l.name) = ?)", strings.ToLower(q.Label))
	}
	if q.DueBefore != nil {
		query = query.Where("tasks.due_date <= ?", *q.DueBefore)
	}
	if q.DueAfter != nil {
		query = query.Where("tasks.due_date >= ?", *q.DueAfter)
	}
	return query
}

// lifecycleFilters hides done tasks, archived tasks and the tasks of
// archived projects unless q asks for them. A status asked for by name is
// always shown.
func lifecycleFilters(query *gorm.DB, q SearchQuery) *gorm.DB {
	if len(q.Statuses) > 0 {
		return query
	}
	if !q.IncludeArchived {
		query = query.Where("tasks.status <> ?", "archived").
			Where("NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = tasks.project_id AND p.archived_at IS NOT NULL)")
	}
	if !q.IncludeDone {
		query = query.Where("NOT EXISTS (SELECT 1 FROM columns c WHERE c.project_id = tasks.project_id AND c.slug = tasks.status AND c.is_final)")
	}
	return query
}

// searchOrder is the ORDER BY of a sort name; unknown names sort by update.
func searchOrder(sort string) string {
	switch sort {
	case SortDue:
		return "CASE WHEN tasks.due_date IS NULL THEN 1 ELSE 0 END, tasks.due_date asc"
	case SortPriority:
		return buildOrderClause("priority", "desc")
	case SortCreated:
		return "tasks.created_at desc"
	case SortNumber:
		return "tasks.project_id asc, tasks.number asc"
	default:
		return "tasks.updated_at desc"
	}
}

// escapeLike escapes the LIKE wildcards in a search word.
func escapeLike(word string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(word)
}

// Response renders a task record as the routes answer it.
func Response(record TaskRecord) TaskResponse {
	return recordResponse(record)
}
