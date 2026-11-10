package mcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// Text limits of the read formats.
const (
	conciseDescription  = 300
	detailedDescription = 20_000
	commentLimit        = 4_000
	dateLayout          = "2006-01-02"
)

// errDate is the refusal for a date in neither accepted format.
var errDate = errors.New("dates are 2026-11-30 or RFC 3339 such as 2026-11-30T17:00:00+01:00")

// parseDate reads a date or an RFC 3339 time; a bare date is midnight UTC.
func parseDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	if t, err := time.Parse(dateLayout, value); err == nil {
		return &t, nil
	}
	return nil, errDate
}

// stamp renders a time in RFC 3339 UTC.
func stamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// day renders an optional date: the date alone at midnight UTC, else RFC
// 3339.
func day(t *time.Time) string {
	if t == nil {
		return ""
	}
	u := t.UTC()
	if u.Hour() == 0 && u.Minute() == 0 && u.Second() == 0 {
		return u.Format(dateLayout)
	}
	return stamp(u)
}

// projectInfo is what a task rendering needs from its project.
type projectInfo struct {
	Slug string
	Name string
}

// projectsOf loads the slug and name of the projects of tasks.
func (t *toolset) projectsOf(ctx context.Context, tasks []task.TaskResponse) map[int64]projectInfo {
	ids := make([]int64, 0, len(tasks))
	for _, tk := range tasks {
		ids = append(ids, tk.ProjectID)
	}
	var rows []schemas.Project
	out := map[int64]projectInfo{}
	if len(ids) == 0 {
		return out
	}
	t.domain.DB.WithContext(ctx).Select("id", "slug", "name").Where("id IN ?", ids).Find(&rows)
	for _, p := range rows {
		out[p.ID] = projectInfo{Slug: p.Slug, Name: p.Name}
	}
	return out
}

// taskURL is the web page of a task.
func (t *toolset) taskURL(tk task.TaskResponse) string {
	return fmt.Sprintf("%s/dashboard/workspace/%d/project/%d/task/%d", t.appURL, tk.WorkspaceID, tk.ProjectID, tk.ID)
}

// projectURL is the board of a project.
func (t *toolset) projectURL(workspaceID, projectID int64) string {
	return fmt.Sprintf("%s/dashboard/workspace/%d/project/%d/board", t.appURL, workspaceID, projectID)
}

// ref is a task's reference: the project key, a dash and the number. Branch
// names and pull requests carrying it link to the task.
func ref(slug string, number int) string {
	return slug + "-" + strconv.Itoa(number)
}

// assignee renders a task's assignee.
func assignee(tk task.TaskResponse) *person {
	if tk.ActorID == nil {
		return nil
	}
	return &person{UserID: *tk.ActorID, Name: tk.ActorName}
}

// labelNames lists a task's label names.
func labelNames(labels []task.TaskLabelResponse) []string {
	names := make([]string, 0, len(labels))
	for _, l := range labels {
		names = append(names, l.Name)
	}
	return names
}

// subtasks renders subtask progress, empty without subtasks.
func subtasks(tk task.TaskResponse) string {
	if tk.SubtaskTotal == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", tk.SubtaskDone, tk.SubtaskTotal)
}

// summarize renders one task of a list.
func (t *toolset) summarize(tk task.TaskResponse, p projectInfo, full bool) taskSummary {
	s := taskSummary{
		TaskID: tk.ID, Ref: ref(p.Slug, tk.Number), ProjectID: tk.ProjectID, Title: tk.Title, Status: tk.Status,
		Priority: tk.Priority, Assignee: assignee(tk), Due: day(tk.DueDate), Labels: labelNames(tk.Labels),
		UpdatedAt: stamp(tk.UpdatedAt),
	}
	if full {
		description, _ := mcpkit.Clip(tk.Description, conciseDescription)
		s.taskDetail = taskDetail{
			WorkspaceID: tk.WorkspaceID, Description: description, Start: day(tk.StartDate),
			CreatedAt: stamp(tk.CreatedAt), Subtasks: subtasks(tk), URL: t.taskURL(tk),
		}
	}
	return s
}

// full renders one task with everything a model may need, comments aside.
func (t *toolset) full(tk task.TaskResponse, p projectInfo) taskOutput {
	description, cut := mcpkit.Clip(tk.Description, detailedDescription)
	return taskOutput{
		taskHead: taskHead{
			TaskID: tk.ID, Ref: ref(p.Slug, tk.Number), WorkspaceID: tk.WorkspaceID, ProjectID: tk.ProjectID,
			ProjectName: p.Name, URL: t.taskURL(tk), CreatedAt: stamp(tk.CreatedAt), UpdatedAt: stamp(tk.UpdatedAt),
		},
		taskBody: taskBody{
			Title: tk.Title, Status: tk.Status, Priority: tk.Priority, Assignee: assignee(tk), Due: day(tk.DueDate),
			Start: day(tk.StartDate), Labels: labelNames(tk.Labels), Description: description,
			DescriptionTruncated: cut, Subtasks: subtasks(tk),
		},
	}
}

// fullOf renders tk after loading its project.
func (t *toolset) fullOf(ctx context.Context, tk task.TaskResponse) taskOutput {
	return t.full(tk, t.projectsOf(ctx, []task.TaskResponse{tk})[tk.ProjectID])
}
