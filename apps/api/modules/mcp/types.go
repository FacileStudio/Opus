package mcp

// workspacesInput is the input of opus_list_workspaces.
type workspacesInput struct{}

// person is a member of a workspace, or the assignee or author of a task.
type person struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

// labelSummary is a label a task of the workspace can carry.
type labelSummary struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// workspaceSummary is one workspace the user belongs to.
type workspaceSummary struct {
	WorkspaceID  int64          `json:"workspace_id"`
	Name         string         `json:"name"`
	Personal     bool           `json:"personal"`
	Role         string         `json:"role"`
	MemberCount  int            `json:"member_count"`
	Members      []person       `json:"members"`
	Labels       []labelSummary `json:"labels"`
	ManagedByIdP bool           `json:"members_managed_in_registre,omitempty"`
}

// workspacesOutput is the output of opus_list_workspaces.
type workspacesOutput struct {
	Workspaces []workspaceSummary `json:"workspaces"`
	Note       string             `json:"note,omitempty"`
}

// projectsInput is the input of opus_list_projects.
type projectsInput struct {
	WorkspaceID     int64  `json:"workspace_id,omitempty" jsonschema:"the workspace, from opus_list_workspaces. Omit for every workspace"`
	Query           string `json:"query,omitempty" jsonschema:"words that must all appear in the project name or description"`
	IncludeArchived bool   `json:"include_archived,omitempty" jsonschema:"also list archived projects"`
	Limit           int    `json:"limit,omitempty" jsonschema:"projects per page, 1 to 100. Default 50"`
	Cursor          string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
	ResponseFormat  string `json:"response_format,omitempty" jsonschema:"concise (default) or detailed, which adds the description, column names and the web link"`
}

// boardColumn is one status of a project's board.
type boardColumn struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Final  bool   `json:"final,omitempty"`
}

// projectSummary is one project.
type projectSummary struct {
	ProjectID   int64    `json:"project_id"`
	WorkspaceID int64    `json:"workspace_id"`
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	Archived    bool     `json:"archived,omitempty"`
	Statuses    []string `json:"statuses"`
	TotalTasks  int      `json:"total_tasks"`
	PercentDone int      `json:"percent_done"`
	projectDetail
}

// projectDetail is what the detailed format adds to a project.
type projectDetail struct {
	Description string        `json:"description,omitempty"`
	Icon        string        `json:"icon,omitempty"`
	Columns     []boardColumn `json:"columns,omitempty"`
	CreatedAt   string        `json:"created_at,omitempty"`
	URL         string        `json:"url,omitempty"`
}

// projectsOutput is the output of opus_list_projects.
type projectsOutput struct {
	Projects   []projectSummary `json:"projects"`
	Total      int              `json:"total"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

// searchInput is the input of opus_search_tasks.
type searchInput struct {
	WorkspaceID     int64    `json:"workspace_id,omitempty" jsonschema:"only this workspace, from opus_list_workspaces"`
	ProjectID       int64    `json:"project_id,omitempty" jsonschema:"only this project, from opus_list_projects"`
	Query           string   `json:"query,omitempty" jsonschema:"words that must all appear in the title or description. Omit to list"`
	Statuses        []string `json:"statuses,omitempty" jsonschema:"only these statuses, such as to-do or in-progress; see a project's statuses. Naming archived or a final status shows those tasks too"`
	Priority        string   `json:"priority,omitempty" jsonschema:"only this priority: no-priority, low, medium, high or urgent"`
	Assignee        string   `json:"assignee,omitempty" jsonschema:"me, none for unassigned tasks, or a user_id from opus_list_workspaces"`
	Label           string   `json:"label,omitempty" jsonschema:"only tasks carrying this label name"`
	DueBefore       string   `json:"due_before,omitempty" jsonschema:"due on or before this date, 2026-11-30 or RFC 3339"`
	DueAfter        string   `json:"due_after,omitempty" jsonschema:"due on or after this date, 2026-11-01 or RFC 3339"`
	IncludeDone     bool     `json:"include_done,omitempty" jsonschema:"also list tasks in a final status such as done. Default false"`
	IncludeArchived bool     `json:"include_archived,omitempty" jsonschema:"also list archived tasks and the tasks of archived projects"`
	Sort            string   `json:"sort,omitempty" jsonschema:"updated (default, newest first), due (soonest first), priority (urgent first), created or number"`
	Limit           int      `json:"limit,omitempty" jsonschema:"tasks per page, 1 to 100. Default 30"`
	Cursor          string   `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page, with the same filters"`
	ResponseFormat  string   `json:"response_format,omitempty" jsonschema:"concise (default) or detailed, which adds a description excerpt, dates and the web link"`
}

// taskSummary is one task of a list.
type taskSummary struct {
	TaskID    int64    `json:"task_id"`
	Ref       string   `json:"ref"`
	ProjectID int64    `json:"project_id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Priority  string   `json:"priority"`
	Assignee  *person  `json:"assignee,omitempty"`
	Due       string   `json:"due,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	UpdatedAt string   `json:"updated_at"`
	taskDetail
}

// taskDetail is what the detailed format adds to a task summary.
type taskDetail struct {
	WorkspaceID int64  `json:"workspace_id,omitempty"`
	Description string `json:"description,omitempty"`
	Start       string `json:"start,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	Subtasks    string `json:"subtasks_done,omitempty"`
	URL         string `json:"url,omitempty"`
}

// searchOutput is the output of opus_search_tasks.
type searchOutput struct {
	Tasks      []taskSummary `json:"tasks"`
	Total      int64         `json:"total"`
	NextCursor string        `json:"next_cursor,omitempty"`
	Note       string        `json:"note,omitempty"`
}

// readInput is the input of opus_read_task.
type readInput struct {
	TaskID         int64  `json:"task_id,omitempty" jsonschema:"the task, from opus_search_tasks"`
	Ref            string `json:"ref,omitempty" jsonschema:"the task's reference instead of task_id, such as opus-42 (project key, then number)"`
	CommentsCursor string `json:"comments_cursor,omitempty" jsonschema:"comments_next_cursor from the previous call, for older comments"`
}

// comment is one comment on a task.
type comment struct {
	CommentID int64   `json:"comment_id"`
	Author    *person `json:"author,omitempty"`
	Text      string  `json:"text"`
	Truncated bool    `json:"text_truncated,omitempty"`
	CreatedAt string  `json:"created_at"`
	Edited    bool    `json:"edited,omitempty"`
}

// taskOutput is one task in full, the output of opus_read_task and of the
// writes.
type taskOutput struct {
	taskHead
	taskBody
	taskThread
}

// taskHead is where a task lives and when it changed.
type taskHead struct {
	TaskID      int64  `json:"task_id"`
	Ref         string `json:"ref"`
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id"`
	ProjectName string `json:"project_name"`
	URL         string `json:"url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// taskBody is what a task says.
type taskBody struct {
	Title                string   `json:"title"`
	Status               string   `json:"status"`
	Priority             string   `json:"priority"`
	Assignee             *person  `json:"assignee,omitempty"`
	Due                  string   `json:"due,omitempty"`
	Start                string   `json:"start,omitempty"`
	Labels               []string `json:"labels,omitempty"`
	Description          string   `json:"description,omitempty"`
	DescriptionTruncated bool     `json:"description_truncated,omitempty"`
	Subtasks             string   `json:"subtasks_done,omitempty"`
}

// taskThread is one page of a task's comments.
type taskThread struct {
	Comments           []comment `json:"comments,omitempty"`
	CommentsTotal      int       `json:"comments_total,omitempty"`
	CommentsNextCursor string    `json:"comments_next_cursor,omitempty"`
}

// createTaskInput is the input of opus_create_task.
type createTaskInput struct {
	ProjectID   int64    `json:"project_id" jsonschema:"the project, from opus_list_projects"`
	Title       string   `json:"title" jsonschema:"the task title"`
	Description string   `json:"description,omitempty" jsonschema:"details, in Markdown"`
	Status      string   `json:"status,omitempty" jsonschema:"one of the project's statuses. Default to-do"`
	Priority    string   `json:"priority,omitempty" jsonschema:"no-priority, low (default), medium, high or urgent"`
	Assignee    string   `json:"assignee,omitempty" jsonschema:"me, or a user_id from opus_list_workspaces. Omit to leave it unassigned"`
	Due         string   `json:"due,omitempty" jsonschema:"due date, 2026-11-30 or RFC 3339"`
	Start       string   `json:"start,omitempty" jsonschema:"start date, 2026-11-20 or RFC 3339"`
	Labels      []string `json:"labels,omitempty" jsonschema:"label names from the workspace's labels in opus_list_workspaces"`
}

// updateTaskInput is the input of opus_update_task. Omitted fields keep
// their value.
type updateTaskInput struct {
	TaskID      int64     `json:"task_id" jsonschema:"the task, from opus_search_tasks"`
	Title       *string   `json:"title,omitempty" jsonschema:"a new title"`
	Description *string   `json:"description,omitempty" jsonschema:"a new description in Markdown, replacing the old one; empty clears it"`
	Status      *string   `json:"status,omitempty" jsonschema:"one of the project's statuses, such as in-progress or done"`
	Priority    *string   `json:"priority,omitempty" jsonschema:"no-priority, low, medium, high or urgent"`
	Assignee    *string   `json:"assignee,omitempty" jsonschema:"me, a user_id from opus_list_workspaces, or none to unassign"`
	Due         *string   `json:"due,omitempty" jsonschema:"a new due date, 2026-11-30 or RFC 3339; empty clears it"`
	Start       *string   `json:"start,omitempty" jsonschema:"a new start date; empty clears it"`
	Labels      *[]string `json:"labels,omitempty" jsonschema:"the full new set of label names; empty removes every label"`
	ProjectID   *int64    `json:"project_id,omitempty" jsonschema:"move the task to this project of the same workspace"`
}

// writeOutput reports a created or updated task.
type writeOutput struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Task    taskOutput `json:"task"`
}

// commentInput is the input of opus_comment_task.
type commentInput struct {
	TaskID  int64  `json:"task_id" jsonschema:"the task, from opus_search_tasks"`
	Comment string `json:"comment" jsonschema:"the comment, in Markdown. It is posted under the user's name"`
}

// commentOutput reports a posted comment.
type commentOutput struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	TaskID  int64   `json:"task_id"`
	Comment comment `json:"comment"`
}

// createProjectInput is the input of opus_create_project.
type createProjectInput struct {
	WorkspaceID int64  `json:"workspace_id,omitempty" jsonschema:"the workspace, from opus_list_workspaces. Omit for the user's personal workspace"`
	Name        string `json:"name" jsonschema:"the project name"`
	Description string `json:"description,omitempty" jsonschema:"what the project is about"`
	Key         string `json:"key,omitempty" jsonschema:"short lowercase key that prefixes task refs, such as web for web-12. Default derived from the name"`
}

// updateProjectInput is the input of opus_update_project. Omitted fields
// keep their value.
type updateProjectInput struct {
	ProjectID   int64   `json:"project_id" jsonschema:"the project, from opus_list_projects"`
	Name        *string `json:"name,omitempty" jsonschema:"a new name"`
	Description *string `json:"description,omitempty" jsonschema:"a new description; empty clears it"`
	Archived    *bool   `json:"archived,omitempty" jsonschema:"true archives the project, false restores it"`
}

// projectOutput reports a created or updated project.
type projectOutput struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Project projectSummary `json:"project"`
}

// taskRef names one task.
type taskRef struct {
	TaskID int64 `json:"task_id" jsonschema:"the task, from opus_search_tasks"`
}

// projectRef names one project.
type projectRef struct {
	ProjectID int64 `json:"project_id" jsonschema:"the project, from opus_list_projects"`
}

// actionOutput reports a delete, or why it did not happen yet.
type actionOutput struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	TaskID     int64  `json:"task_id,omitempty"`
	ProjectID  int64  `json:"project_id,omitempty"`
	ConfirmURL string `json:"confirm_url,omitempty"`
}
