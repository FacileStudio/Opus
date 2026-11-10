package mcp

import (
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// call runs a tool and decodes its result into out.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	mcptest.Decode(t, mcptest.Call(t, session, tool, args), out)
}

// refused runs a tool, expects an error result and returns its text.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must fail, got %+v", tool, result.StructuredContent)
	}
	return mcptest.Text(result)
}

// TestTaskLifecycle drives a task from creation to done the way an agent
// would: create with an assignee and a label, find it, comment, close it.
func TestTaskLifecycle(t *testing.T) {
	f := newFixture(t)
	s := f.session(t)
	var created writeOutput
	call(t, s, toolCreateTask, map[string]any{
		"project_id": f.project, "title": "Fix the login page", "description": "It **crashes**.",
		"assignee": strconv.FormatInt(f.teammate, 10), "due": "2026-12-01", "labels": []string{"BUG"}, "priority": "high",
	}, &created)
	tk := created.Task
	if tk.Ref != "web-1" || tk.Assignee == nil || tk.Assignee.UserID != f.teammate || tk.Due != "2026-12-01" ||
		len(tk.Labels) != 1 || tk.Labels[0] != "bug" || tk.Status != "to-do" {
		t.Fatalf("created task = %+v", tk)
	}

	var found searchOutput
	call(t, s, toolSearchTasks, map[string]any{"query": "login", "label": "bug", "assignee": strconv.FormatInt(f.teammate, 10)}, &found)
	if found.Total != 1 || found.Tasks[0].TaskID != tk.TaskID {
		t.Fatalf("search = %+v", found)
	}

	var commented commentOutput
	call(t, s, toolCommentTask, map[string]any{"task_id": tk.TaskID, "comment": "On it"}, &commented)
	var read taskOutput
	call(t, s, toolReadTask, map[string]any{"ref": "WEB-1"}, &read)
	if read.TaskID != tk.TaskID || read.CommentsTotal != 1 || read.Comments[0].Text != "On it" || read.Comments[0].Author.UserID != f.owner {
		t.Fatalf("read = %+v", read)
	}

	var updated writeOutput
	call(t, s, toolUpdateTask, map[string]any{"task_id": tk.TaskID, "status": "done", "assignee": "none", "labels": []string{}, "due": ""}, &updated)
	if updated.Task.Status != "done" || updated.Task.Assignee != nil || len(updated.Task.Labels) != 0 || updated.Task.Due != "" ||
		updated.Task.Title != "Fix the login page" || updated.Task.Priority != "high" {
		t.Fatalf("updated = %+v", updated.Task)
	}
	call(t, s, toolSearchTasks, map[string]any{"query": "login"}, &found)
	if found.Total != 0 {
		t.Fatalf("a done task is hidden by default, got %+v", found)
	}
	call(t, s, toolSearchTasks, map[string]any{"query": "login", "include_done": true}, &found)
	if found.Total != 1 {
		t.Fatalf("include_done lists the done task, got %+v", found)
	}
}

// TestBadInputIsActionable checks refusals tell the model what to do.
func TestBadInputIsActionable(t *testing.T) {
	f := newFixture(t)
	s := f.session(t)
	for _, c := range []struct {
		tool string
		args map[string]any
		want string
	}{
		{toolCreateTask, map[string]any{"project_id": f.project, "title": "x", "labels": []string{"nope"}}, "opus_list_workspaces"},
		{toolCreateTask, map[string]any{"project_id": f.project, "title": "x", "assignee": strconv.FormatInt(f.stranger, 10)}, "not a member"},
		{toolCreateTask, map[string]any{"project_id": f.project, "title": "x", "status": "shipped"}, "Invalid status"},
		{toolCreateTask, map[string]any{"project_id": f.project, "title": "x", "due": "tomorrow"}, "2026-11-30"},
		{toolSearchTasks, map[string]any{"cursor": "garbage!"}, "cursor"},
		{toolReadTask, map[string]any{"task_id": 999_999}, "opus_search_tasks"},
	} {
		if got := refused(t, s, c.tool, c.args); !strings.Contains(got, c.want) {
			t.Errorf("%s %v: %q lacks %q", c.tool, c.args, got, c.want)
		}
	}
}

// TestOtherWorkspacesStayInvisible checks a tool sees only what the web app
// shows the person: a task in a workspace they are not in is never listed,
// read, changed or offered for deletion.
func TestOtherWorkspacesStayInvisible(t *testing.T) {
	f := newFixture(t)
	secret := f.task(t, f.foreign, f.stranger, "Secret plan")
	s := f.session(t)
	var found searchOutput
	call(t, s, toolSearchTasks, map[string]any{"query": "secret"}, &found)
	if found.Total != 0 {
		t.Fatalf("search leaked a foreign task: %+v", found)
	}
	var projects projectsOutput
	call(t, s, toolListProjects, map[string]any{}, &projects)
	for _, p := range projects.Projects {
		if p.ProjectID == f.foreign {
			t.Fatalf("list_projects leaked a foreign project")
		}
	}
	refused(t, s, toolReadTask, map[string]any{"task_id": secret.ID})
	refused(t, s, toolReadTask, map[string]any{"ref": "secret-1"})
	refused(t, s, toolUpdateTask, map[string]any{"task_id": secret.ID, "title": "mine"})
	refused(t, s, toolCommentTask, map[string]any{"task_id": secret.ID, "comment": "hi"})
	refused(t, s, toolCreateTask, map[string]any{"project_id": f.foreign, "title": "intrusion"})
	refused(t, s, toolDeleteTask, map[string]any{"task_id": secret.ID})
	refused(t, s, toolDeleteProject, map[string]any{"project_id": f.foreign})
	refused(t, s, toolSearchTasks, map[string]any{"project_id": f.foreign})
	open, _ := f.pending.Open(t.Context(), strconv.FormatInt(f.owner, 10))
	if len(open) != 0 || !f.exists(&schemas.Task{}, secret.ID) {
		t.Fatalf("a refused delete must leave no pending action and the task, got %d actions", len(open))
	}
	var check schemas.Task
	f.db.First(&check, secret.ID)
	if check.Title != "Secret plan" {
		t.Fatalf("the foreign task changed: %q", check.Title)
	}
}

// TestProjectsAndWorkspaces covers the project tools and the workspace list.
func TestProjectsAndWorkspaces(t *testing.T) {
	f := newFixture(t)
	s := f.session(t)
	var spaces workspacesOutput
	call(t, s, toolListWorkspaces, map[string]any{}, &spaces)
	if len(spaces.Workspaces) != 2 || !spaces.Workspaces[0].Personal || spaces.Workspaces[1].Role != "owner" ||
		spaces.Workspaces[1].MemberCount != 2 || len(spaces.Workspaces[1].Labels) != 1 {
		t.Fatalf("workspaces = %+v", spaces)
	}
	var created projectOutput
	call(t, s, toolCreateProject, map[string]any{"name": "Side Quest", "description": "Fun"}, &created)
	if created.Project.WorkspaceID != spaces.Workspaces[0].WorkspaceID || created.Project.Key != "side-quest" ||
		created.Project.Description != "Fun" || len(created.Project.Statuses) != 6 {
		t.Fatalf("created project = %+v", created.Project)
	}
	var archived projectOutput
	call(t, s, toolUpdateProject, map[string]any{"project_id": created.Project.ProjectID, "archived": true, "name": "Side"}, &archived)
	if !archived.Project.Archived || archived.Project.Name != "Side" {
		t.Fatalf("archived project = %+v", archived.Project)
	}
	var listed projectsOutput
	call(t, s, toolListProjects, map[string]any{"query": "side"}, &listed)
	if listed.Total != 0 {
		t.Fatalf("an archived project is hidden by default, got %+v", listed)
	}
	call(t, s, toolListProjects, map[string]any{"query": "side", "include_archived": true}, &listed)
	if listed.Total != 1 {
		t.Fatalf("include_archived lists it, got %+v", listed)
	}
}

// TestMoveKeepsTheTask checks a project_id update moves the task within the
// workspace and refuses another workspace.
func TestMoveKeepsTheTask(t *testing.T) {
	f := newFixture(t)
	other := f.newProject(t, f.studio, "Api", "api")
	tk := f.task(t, f.project, f.owner, "Move me")
	s := f.session(t)
	var moved writeOutput
	call(t, s, toolUpdateTask, map[string]any{"task_id": tk.ID, "project_id": other, "status": "in-progress"}, &moved)
	if moved.Task.ProjectID != other || moved.Task.Status != "in-progress" || !strings.HasPrefix(moved.Task.Ref, "api-") {
		t.Fatalf("moved = %+v", moved.Task)
	}
	refused(t, s, toolUpdateTask, map[string]any{"task_id": tk.ID, "project_id": f.foreign})
}

// TestPaging checks the cursor walks every task once.
func TestPaging(t *testing.T) {
	f := newFixture(t)
	for i := range 5 {
		f.task(t, f.project, f.owner, "Item "+strconv.Itoa(i))
	}
	s := f.session(t)
	seen := map[int64]bool{}
	cursor := ""
	for range 10 {
		var page searchOutput
		call(t, s, toolSearchTasks, map[string]any{"limit": 2, "cursor": cursor, "sort": "number"}, &page)
		for _, tk := range page.Tasks {
			seen[tk.TaskID] = true
		}
		if cursor = page.NextCursor; cursor == "" {
			break
		}
	}
	if len(seen) != 5 {
		t.Fatalf("paging saw %d of 5 tasks", len(seen))
	}
}
