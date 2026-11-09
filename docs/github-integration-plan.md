# Opus — GitHub Integration Implementation Plan

## Goal
Add a clean, native GitHub integration to Opus that connects code activity (branches, pull requests) to tasks, automatically updates task workflows, and resolves user attributions with zero configuration via Registre's GitHub OAuth.

## Why (evidence)
- Registre (the Facile suite IdP) recently introduced GitHub OAuth (`read:user user:email`) and emits the `connections.github` claim (`apps/api/internal/social/github.go`).
- Opus users currently copy branch names via `task-properties-sidebar.tsx` (`{slug}-{number}`, shortcut `Ctrl+Shift+G`), but Opus has no webhook listener or repository integration to react when branches are pushed or PRs are opened and merged.
- Prior to the Go migration, legacy TypeScript Opus contained a bloated 3,800-line third-party integration module with heavy external dependencies, which was intentionally dropped during the port to build a clean, minimalist Facile-native implementation.

## Approach
Separate **Authentication / Identity** from **Repository Automation**:
1. **Identity via Registre**: Registre maps the developer's Facile user to their GitHub ID (`connections.github`). Porte passes this in `porte.Claims.Extra["connections"]`. Opus stores `users.github_id` during OIDC upsert. No user-level GitHub tokens or secondary user auth required.
2. **Repository Automation via GitHub App**: Workspaces/repositories install a standard GitHub App with webhook subscriptions. Webhooks (`pull_request`, `push`) post to `POST /api/integrations/github/webhook`.
3. **Task & Workflow Engine**:
   - Webhook handler verifies `X-Hub-Signature-256`.
   - Branch names, PR titles, and PR bodies are parsed with regex `(?i)\b([a-z0-9_-]+)-(\d+)\b` matching `{project_slug}-{task_number}`.
   - Matched tasks record linked PR metadata in `task_pull_requests`.
   - Automatic state transitions:
     - Branch created / PR opened as draft $\rightarrow$ move to column with slug `in-progress` (and assign author if unassigned).
     - PR marked ready for review $\rightarrow$ move to column with slug `in-review`.
     - PR merged to default branch $\rightarrow$ move to column with `is_final = true` (Done).
4. **UI**:
   - Task properties sidebar displays a "Linked Pull Requests" list with live status badges (`Draft`, `Open`, `Merged`, `Closed`), repo name, PR number, and direct link.

**Checked against**:
- `[migrations]`: Goose sequential migration `apps/api/migrations/00007_github_integration.sql` with `-- +goose Up` and `-- +goose Down`.
- `[auth/porte]`: Webhook receiver is public/signature-verified (`X-Hub-Signature-256`), bypassing cookie auth; OIDC connection claims read through `porte.Config.ExtraClaims`.
- `[filet]`: Code adheres to cognitive complexity $\le 15$, statement counts $\le 40$, no inline comments.
- `[module-path]`: `github.com/FacileStudio/Opus/apps/api`.

---

## Phased Breakdown

### Phase 1: The Core Loop (In Scope for this release)
- **Database**:
  - Add `github_id TEXT` column to `users` table.
  - Create `project_github_repos` table for repository mapping.
  - Create `task_pull_requests` table for tracking linked PRs.
- **Backend API**:
  - Update `internal/env/env.go` to include `connections` in `porte.Config.ExtraClaims` and read `GITHUB_WEBHOOK_SECRET`.
  - Update `modules/auth/users.go` to extract `connections.github` and persist `user.github_id`.
  - Create `modules/github` containing webhook handler, signature verification, PR/branch parser, and task state machine.
  - Wire route `POST /api/integrations/github/webhook` in `apps/api/router.go`.
  - Expose `GET /api/project/:projectId/task/:taskId/pull-requests` (or embed in task response).
- **Frontend Web**:
  - Add `PullRequestCard` / `LinkedPullRequests` component in `apps/web/src/components/task/task-properties-sidebar.tsx`.
  - Display PR badges with states: `Draft` (gray), `Open` (blue), `Merged` (purple), `Closed` (red).

### Phase 2: Rich Context & PR Backlinks (Future)
- GitHub App bot posts backlink comment on GitHub PR with Opus task link and status.
- Show PR review approvals and CI check status directly inside Opus.

### Phase 3: Facile Suite Event Hooks (Future)
- Emit internal event on PR merge to auto-stop running Sablier timer on that task.
- Forward deployment preview links (Dokploy/Nacelle) to task view.

---

## Technical Specifications (Phase 1)

### 1. Database Schema (`apps/api/migrations/00007_github_integration.sql`)

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS github_id TEXT;
CREATE INDEX IF NOT EXISTS idx_users_github_id ON users (github_id) WHERE github_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS project_github_repos (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, repo_full_name)
);
CREATE INDEX IF NOT EXISTS idx_project_github_repos_name ON project_github_repos (repo_full_name);

CREATE TABLE IF NOT EXISTS task_pull_requests (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    pr_number INT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    state TEXT NOT NULL, -- 'draft', 'open', 'merged', 'closed'
    head_branch TEXT NOT NULL DEFAULT '',
    base_branch TEXT NOT NULL DEFAULT '',
    author_github_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, repo_full_name, pr_number)
);
CREATE INDEX IF NOT EXISTS idx_task_pull_requests_task ON task_pull_requests (task_id);

-- +goose Down
DROP TABLE IF EXISTS task_pull_requests;
DROP TABLE IF EXISTS project_github_repos;
DROP INDEX IF EXISTS idx_users_github_id;
ALTER TABLE users DROP COLUMN IF EXISTS github_id;
```

### 2. GORM Models (`apps/api/schemas/models.go`)

```go
type ProjectGitHubRepo struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID     int64     `gorm:"column:project_id;not null;index"`
	RepoFullName  string    `gorm:"column:repo_full_name;not null"`
	DefaultBranch string    `gorm:"column:default_branch;not null;default:'main'"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
}

type TaskPullRequest struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID         int64     `gorm:"column:task_id;not null;index"`
	RepoFullName   string    `gorm:"column:repo_full_name;not null"`
	PRNumber       int       `gorm:"column:pr_number;not null"`
	Title          string    `gorm:"column:title;not null"`
	URL            string    `gorm:"column:url;not null"`
	State          string    `gorm:"column:state;not null"`
	HeadBranch     string    `gorm:"column:head_branch;not null;default:''"`
	BaseBranch     string    `gorm:"column:base_branch;not null;default:''"`
	AuthorGitHubID *string   `gorm:"column:author_github_id"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}
```

### 3. Identity Linking (`apps/api/modules/auth/users.go`)

In `UpsertFromOIDC`:
```go
if raw, ok := claims.Extra["connections"]; ok && len(raw) > 0 {
    var conns struct {
        GitHub string `json:"github"`
    }
    if err := json.Unmarshal(raw, &conns); err == nil && conns.GitHub != "" {
        updates["github_id"] = conns.GitHub
    }
}
```

### 4. Webhook Handler (`apps/api/modules/github/webhook.go`)
- Route: `POST /api/integrations/github/webhook`
- Verification: HMAC-SHA256 of payload with `GITHUB_WEBHOOK_SECRET` header `X-Hub-Signature-256`.
- Events:
  - `pull_request`: actions `opened`, `reopened`, `closed`, `ready_for_review`, `converted_to_draft`, `synchronize`.
  - `push`: commit messages and ref parsed for task refs.
- Task Key Matching:
  - Regex: `(?i)(?:^|[^a-z0-9_-])(?P<slug>[a-z0-9_-]+)-(?P<num>\d+)\b`
  - Matches project by `slug` (case-insensitive) and task by `number`.
- State Automation Rules:
  - If PR state is draft / opened: find column with slug `in-progress` (or name matching `In Progress`), update `task.column_id`, update `task.status = column.slug`.
  - If author matches a user via `github_id`, and `task.actor_id` is null, assign `task.actor_id = user.id`.
  - If PR is ready for review: find column with slug `in-review` (or name matching `In Review`), update column.
  - If PR is merged (`pr.merged == true` and `base_branch == default_branch`): find column where `is_final == true` (Done), update column.
  - Broadcast realtime task update event (`realtime.EventTaskUpdated`).

### 5. Frontend (`apps/web`)
- Fetch task pull requests via endpoint or task detail response.
- Render in `TaskPropertiesSidebar`:
  - Section title: "Pull requests"
  - For each linked PR:
    - Status icon/pill:
      - `draft`: neutral badge "Draft"
      - `open`: info badge "Open"
      - `merged`: success/purple badge "Merged"
      - `closed`: destructive badge "Closed"
    - PR title, `#123`, repository name.
    - Clickable link opening in new tab.

---

### Antenne Event Bus Integration (Suite-Native Ingress)
In addition to direct GitHub webhooks, Opus registers with the Antenne Pool WebSocket bus for `git.pull_request` and `git.push` channels. When GitHub fires webhooks to Antenne (`https://antenne.facile.studio/webhook/git`), Antenne fans out the event across the pool bus to connected subscribers, eliminating duplicate webhook configuration in GitHub across suite apps.

---

## File Inventory

| Path | Action | Role |
|---|---|---|
| `apps/api/migrations/00007_github_integration.sql` | New | Schema migration for GitHub tables and `users.github_id` |
| `apps/api/schemas/models.go` | Modify | Add GORM structs `ProjectGitHubRepo` and `TaskPullRequest` |
| `apps/api/internal/env/env.go` | Modify | Add `GitHubWebhookSecret` and configure `ExtraClaims: []string{"connections"}` in Porte |
| `apps/api/modules/auth/users.go` | Modify | Extract `connections.github` in `UpsertFromOIDC` |
| `apps/api/modules/auth/users_test.go` | Modify | Test GitHub ID persistence during OIDC upsert |
| `apps/api/modules/github/webhook.go` | New | Signature check, payload parsing, event routing |
| `apps/api/modules/github/linker.go` | New | Regex parsing, task lookup, state transition logic |
| `apps/api/modules/github/types.go` | New | Webhook payload structs and PR models |
| `apps/api/modules/github/router.go` | New | Router registration for webhook and task PR list |
| `apps/api/modules/github/webhook_test.go` | New | Comprehensive tests with simulated GitHub webhook payloads |
| `apps/api/router.go` | Modify | Mount GitHub module under `/api/integrations/github` |
| `apps/web/src/components/task/task-properties-sidebar.tsx` | Modify | Add Linked Pull Requests display |
| `CHANGELOG.md` | Modify | Add entry under `[Unreleased]` |

---

## Exit Criteria
1. `apps/api/migrations/00007_github_integration.sql` executes cleanly on startup.
2. `cd apps/api && go test ./...` passes 100% including all new GitHub tests.
3. `filet check` runs with 0 errors.
4. `pnpm --filter @opus/web typecheck` passes with 0 errors.
5. `pnpm --filter @opus/web test` passes.
6. A test webhook payload with `opus-42` correctly transitions task #42 and records the PR.
7. Ready to tag and ship via `/ship`.
