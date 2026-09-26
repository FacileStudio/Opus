# Opus backend to Go port plan

Cold-start handoff. A fresh session with no prior conversation can pick up any phase from
this document. Branch `feat/go-backend`. Commits with `git commit --no-verify` (lefthook runs
biome and full pnpm build; see Gotchas).

## Goal

Replace the TypeScript/Hono Opus backend with a Go backend at `apps/api`, on the Facile
stack (**tronc + porte + journal + casier + filet**), preserving the React frontend and its
HTTP contract through a strangler-fig transition.

## Why

The suite consolidated the plumbing Opus re-implements (tronc chassis, porte auth, journal
logging) and the roadmap declares Opus "excluded on purpose... being rewritten in Go".
`apps/api-legacy/src/` is ~31k LOC TS across 25 feature modules; better-auth is the suite's
odd-one-out identity engine.

## Approach

Go API at `apps/api` on tronc + porte + journal (+ casier for secrets), implementing the
canonical Facile module pattern (`modules/<feature>/{router,controller,service,types,documentation}.go`).
The old TS API lives at `apps/api-legacy` and keeps running; the Go binary is the single
router (serves the SPA at `/client` + `/api/*`) and proxies unported routes to the legacy
API at `LEGACY_API_URL` (127.0.0.1:1337). Features flip Go-native one module at a time.
The frontend changes once: when auth swaps from better-auth to porte (Phase 1).

## Status: where the port stands right now

- **Phase 0 DONE** (scaffold + Docker single-container dual-run): commits `1c3b48f7`,
  `7cf0d993`, `15a4e3dd`.
- **Phase 1 K1 DONE** (commit `d733de31`): the Go auth seam (porte over portepg with a GORM
  UserStore, `/auth/me`, local `/auth/register` + `/auth/login` answering `{token,user}`) and
  the workspace module (list, members, logo upload/delete/serve).
- **Phase 1 K2 DONE** (commits `bcd49d32` Go, `b5101033` workspace/team web,
  `9f7ad953` auth-surface web, `44032b13` keystone): frontend re-plumbed from better-auth
  to the porte fetch client. Contract lived in `docs/k2-contract.md` (frozen record).
  Summary: workspace CRUD + membership + full invitation lifecycle Go-native (int64
  `Invitation` table, `pending->accepted` inserts a membership); `PATCH /auth/me` +
  `POST /auth/change-password` (porte **v0.3.0** for `Kit.ChangePassword`); the legacy proxy
  authenticates with porte and stamps `X-Opus-User-Id`/`-Email` + HMAC tag (`AUTH_BRIDGE_SECRET`)
  that legacy TS API verifies. Web: cut `apiKey`/`device`/OTP/anonymous/magic-link/plugin SSO,
  porte flow on auth pages, `X-Facile-CSRF` on mutations, session via `/auth/me`.
- **Phase 1.5 DONE (Clean baseline gate)**: All 24 filet warnings and info notices
  eliminated. Composer pattern implemented for auth and workspace routers; management.go
  and invitations.go split under 300 lines; `filet check .` outputs 0 errors, 0 warnings.
- **Phase 2 DONE (Core domain breakdown)**: All 7 domain modules (`project`, `column`,
  `task`, `taskrelation`, `label`, `comment`, `activity`) built, tested, and mounted
  ahead of the legacy proxy. Router logic cleanly extracted to `apps/api/router.go`.
  Workspace boundaries enforced on all operations via `workspace_members`. Test suites
  pass on in-memory SQLite (`go test ./...`). `filet check .` confirms 0 errors and 0
  warnings across 85 files (12,114 lines of Go).
- **Phase 3 DONE (Collaboration breakdown)**: All 5 collaboration modules (`sharelink`,
  `timeentry`, `workflowrule`, `notification`, `search`) built, tested, and mounted
  in `apps/api/router.go`. Strict 5-file architecture respected. Tests pass against
  in-memory SQLite. `filet check .` reports 0 errors and 0 warnings across 116 files
  (15,936 lines of Go).
- **Phase 4 DONE (Events, realtime, and scheduler)**:
  - Antenne event publisher in `internal/events` with suite standard `enveloppe.Event` structure.
  - WebSocket hub in `internal/realtime` mounted at `/api/ws` and `/api/ws/{projectId}`, with in-memory adapter and Redis pub/sub adapter (`REDIS_URL`), cookie/token authentication, and initiator exclusion.
  - Background maintenance scheduler in `internal/scheduler` executing stale invitation pruning and due date notification alerts.
  - Quality verification: `filet check .` reports 0 errors and 0 warnings across 138 files (18,912 lines of Go). All unit tests pass (`go test ./...`).
- **Phase 6 DONE (Cutover and decommission)**:
  - Native asset & file storage implemented in `apps/api/modules/asset/`, `internal/storage/asset.go`, and `internal/storage/s3.go` supporting local disk and S3 with AWS SigV4 presigned URLs.
  - Baseline Goose migration created in `apps/api/migrations/00001_baseline_opus.sql` and wired in `internal/database/migrations.go` with Postgres session lock and GORM fallback.
  - Legacy proxy eliminated from `apps/api/router.go`; backend is 100% Go-native.
  - Dropped third-party integration settings (GitHub, Gitea, Slack, Discord, Telegram) and custom webhook UI surfaces cleaned up in `apps/web/`.
  - Single-container Dokploy deployment configured in `Dockerfile` and `deploy/opus-entrypoint.sh` serving both SPA (`/client`) and API (`/api`).
  - Quality verification: `filet check .` reports 0 errors and 0 warnings across 147 files (19,946 lines of Go). All unit tests pass (`go test ./...`). SPA builds cleanly (`pnpm --filter @opus/web build`).

### K2 leftovers to remember

- `pnpm install` was needed to re-link `@opus/api` into `apps/web/node_modules` after the
  Phase 0 move; `tsc` reported `TS2307` until then. Pre-existing baseline typecheck errors
  (~19 unique) in `apps/api-legacy` and phase-2 test/UI fixtures are not K2 regressions.
- `docs/k2-contract.md` is a frozen record of the K2 contract. Keep for parity checks.
- Route shadow to remember: Go `GET /invitation/{invitationId}` owns `/api/invitation/*`;
  the legacy `/invitation/pending` and `/invitation/public/:id` are dead/kept. The web
  pending-invites list was re-pointed to `authClient.organization.listUserInvitations()`
  (`apps/web/src/fetchers/invitation/get-pending-invitations.ts`), and the public accept
  page fetcher was rewritten to plain `fetch(getApiUrl("/invitation/public/{id}"))`.
- Settings UI: Developer/API-keys tab and components deleted with the api-key surface.

If this document disagrees with the repo, the repo is the truth and this page is stale;
lint it and push the correction.

## Decisions (baked in: re-open only if something breaks against prod)

- **The Go backend owns its own schema on its own database.** The legacy API still
  derives `DATABASE_URL` from `POSTGRES_*` (the CUID2 `opus` database; Drizzle remains
  the live schema authority for the legacy app). The Go binary is scoped to its own
  `GO_DATABASE_URL` (default `postgresql://.../opus_go`), created by
  `deploy/db/init/01-opus-go.sql` on a fresh volume or `createdb` by hand.
- **Identity is int64, matching the suite and porte.** Opus was the only suite app on
  CUID2 text ids; porte froze `Identity.UserID int64` across its whole contract. The port
  adopts the suite's int64 model. The frontend treats ids as opaque strings, so int64
  serialises unchanged through URL building.
- **Schema ownership of the Go tables is GORM AutoMigrate** (like Sablier), not goose.
  Goose on its own baseline remains a Phase 6 step, from a `pg_dump` of the live DB at
  cutover.
- **Casier = fail-soft.** `CASIER_MACHINE_TOKEN` set -> pull env secrets from Casier at boot
  (`GET /api/projects/{project}/environments/{env}/secrets/export` with machine token);
  on any failure, log and continue with `.env`/env vars.
- **`@opus/api` package name is kept** (directory moved to `apps/api-legacy`, name did not):
  `packages/libs` (`workspace:*`, `AppType` import) and the test suite depend on the name.

## Module architecture standard (the contract every module follows)

Source of truth: the `architecture` block in `filet.yml` at repo root, inherited verbatim
from the Go-family fork source (Sablier's own `filet.yml`), plus Plume and Vision
`apps/api/modules/{auth,clients,verify,documents}` for the composer pattern.

Every package under `apps/api/modules/<feature>/` divides responsibility into exactly 5 files:

| File | Responsibility | Mechanical and architectural rules |
|---|---|---|
| `router.go` | HTTP only: route declarations, middleware binding, URL param parsing, JSON decode, status codes | **Required by filet**. Must use composer pattern. Handlers delegate to `service.controller`. No database access. |
| `controller.go` | Request adaptation: payload validation, parameter checks, DTO transformations, mapping errors to `tronc/errors` | Receives `context.Context` and DTOs from `router.go`. Calls `service`. No direct ORM or DB access. |
| `service.go` | Business logic, persistence, transaction management. Holds `*gorm.DB` and unexported `*Controller` | The **only** file in the module that references `*gorm.DB`. Exports `NewService(db)`. |
| `types.go` | Request and response DTO structs | Pure data structures with `json` tags. No business logic or methods. |
| `documentation.go` | OpenAPI route registry for `/docs` (`var Documentation`) | Declarative route descriptors composed via nested `append` helpers. |
| Concern files (`crud.go`, `members.go`, etc.) | Optional splits for large modules | Used only when a concern exceeds size limits. Must not bypass the 5-file roles or line limits. |

### Mechanical limits enforced by filet

All Go code must strictly pass `filet check .` with **zero errors and zero warnings**:

- **File length**: Maximum 300 lines per file (`limits.fileLines: 300`). Files exceeding 300 lines must be split into logical concern files.
- **Function length**: Maximum 60 lines per function (`limits.funcLines: 60`).
- **Function statements**: Maximum 40 statements per function (`limits.funcStatements: 40`).
- **Cognitive complexity**: Maximum 15 per function (`limits.complexity: 15`).
- **Parameter count**: Maximum 5 parameters per function (`limits.params: 5`). Pack larger argument lists into dedicated parameter structs.
- **Return value count**: Maximum 3 return values (`limits.returns: 3`). Pack multi-value returns into dedicated result structs.
- **Nesting depth**: Maximum 5 levels (`limits.nesting: 5`).
- **Code style bans**: No inline comments, no comments inside function bodies, no TODOs, no package-level mutable state, no `init()` functions.

### The composer pattern in `router.go`

`filet` charges anonymous function literals (closures) to their enclosing top-level function.
A single `RegisterRoutes` function declaring dozens of inline route closures accumulates
statements and cognitive complexity from every closure, quickly triggering `go.func.complexity`
and `go.func.statements` warnings.

To prevent this:
1. `RegisterRoutes` must be a short composer function (~5 to 15 lines, complexity near 0).
2. It groups routes and delegates to top-level helper functions: `register<Resource>Routes(router, service)`.
3. Route handlers inside helper functions do not contain business logic or DB calls; they
   decode the body, call `service.controller.<action>`, and write the JSON response.

Example pattern from Plume:

```go
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
    router.Route("/projects", func(router chi.Router) {
        router.Use(middleware.RequireAuth(authService))
        registerCollectionRoutes(router, service)
        registerItemRoutes(router, service)
    })
}

func registerCollectionRoutes(router chi.Router, service *Service) {
    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        resp, err := service.controller.list(r.Context(), r)
        if err != nil {
            httpjson.WriteError(w, err)
            return
        }
        httpjson.WriteJSON(w, http.StatusOK, resp)
    })
}
```

### Wiring contract in `main.go`

`main.go` registers each module through two standard calls:

```go
projectService := project.NewService(db)
project.RegisterRoutes(api, projectService, authService)
```

`NewService` creates the service, instantiates an unexported controller
(`s.controller = newController(s)`), and returns the service. The controller is never
exported or initialized outside the package.

## Multi-tenancy and workspace boundaries (Facile suite spaces standard)

Opus multi-tenancy follows the Facile suite rules documented in `conventions/facile-spaces.md`:

1. **Workspace ID is a security boundary, not an optional filter.**
   Never write queries that simply filter by `WHERE workspace_id = ?` without verifying that
   the authenticated caller is a member of that workspace. Every domain request must verify
   membership in `workspace_members`.
2. **Instance admins get no automatic bypass.**
   An admin cannot access workspaces they are not explicitly invited to. Membership is the sole key.
3. **No ownerless workspaces (safe concurrency on leave/demotion).**
   Leaving or deleting a workspace requires counting owners and updating within a single transaction
   using row-locking (`SELECT ... FOR UPDATE` on postgres) to avoid TOCTOU races between co-owners.
4. **Active workspace scoping.**
   The active workspace travels via URL path (`/workspace/{workspaceId}/...`) or query parameter
   `?workspace_id=`. Missing or invalid workspace context must fail with `400 Bad Request` or
   `404 Not Found`, never silently fallback to personal or unscoped data.

## Phase 1.5: Clean baseline gate (zero-debt refactor) (DONE)

All 24 filet warnings and info notices were eliminated across the Go backend. `filet check .`
exits 0 with 0 errors and 0 warnings.

### Remediation tasks

1. **Refactor `modules/workspace/router.go` into the composer pattern**:
   - Current: complexity 160, 295 lines, 234 statements.
   - Fix: Break into top-level functions `registerWorkspaceCRUD`, `registerMemberRoutes`,
     `registerLogoRoutes`, and `registerInvitationRoutes`.
2. **Refactor `modules/auth/router.go` into the composer pattern**:
   - Current: complexity 34, 71 lines, 58 statements.
   - Fix: Break into `registerCredentials`, `registerMe`, and `registerPassword`.
3. **Split `modules/workspace/management.go`**:
   - Current: 424 lines (limit 300); `Update` complexity 16.
   - Fix: Split into `crud.go` (workspace create/update/delete) and `members.go` (membership
     queries and role modifications). Simplify `Update` branching.
4. **Split `modules/workspace/invitations.go`**:
   - Current: 305 lines (limit 300).
   - Fix: Split invitation acceptance and public handling into `invitations_public.go`.
5. **Fix return value counts**:
   - `internal/storage/logo.go:71`: `ReadWorkspaceLogo` returns 4 values. Replace with
     `(file *LogoFile, err error)`.
   - `modules/workspace/service.go:129`: `Logo` returns 4 values. Return a struct or reduce to 2.
   - `main.go:68`: `buildAuth` returns 4 values. Bundle into `authComponents` struct.
6. **Fix parameter counts**:
   - `modules/auth/service.go:80`: `Register` takes 6 parameters. Pack into `RegisterParams`.
   - `modules/auth/service.go:131`: `ChangePassword` takes 6 parameters. Pack into `ChangePasswordParams`.
   - `main.go:107`: `buildRouter` takes 7 parameters. Pack dependencies into `routerConfig` struct.
7. **Fix function complexity and resource handling in `main.go` and `internal/env`**:
   - `main.go:183`: `run` complexity 18, 77 lines. Extract setup steps into helper functions.
   - `main.go:237`: Server implements `io.Closer`; ensure proper cleanup/shutdown invocation.
   - `internal/env/env.go:68`: `Load` complexity 17. Extract sub-parsers for Casier and Porte config.

Gate: `go test ./...` passes and `filet check .` reports 0 errors and 0 warnings.

## Phase 2: Core domain breakdown (DONE)

Phase 2 replaces the legacy task and project management systems. Rather than building a
monolithic module, the domain is split into distinct, decoupled packages under `apps/api/modules/`.
Each package strictly adheres to the 5-file architecture (< 300 lines per file).

All 7 modules are fully implemented, tested against in-memory SQLite, mounted ahead of
the legacy proxy in `apps/api/router.go`, and verified clean by `filet check .` (85 files,
12,114 lines, 0 errors, 0 warnings).

### Sub-module matrix

| Module | Endpoints | Responsibilities | Key schemas | Status |
|---|---|---|---|---|
| `modules/project` | `/project/*` | Project CRUD, project settings, view configurations | `Project` | Done (tests pass) |
| `modules/column` | `/column/*` | Kanban columns, column ordering and reorder | `Column` | Done (tests pass) |
| `modules/task` | `/task/*` | Task CRUD, filters, assignment, priority, due dates, task counter/numbering, bulk updates | `Task` | Done (tests pass) |
| `modules/taskrelation` | `/task-relation/*` | Task dependencies (subtask, blocks, related) | `TaskRelation` | Done (tests pass) |
| `modules/label` | `/label/*` | Workspace labels, task attachments, color tagging | `Label` | Done (tests pass) |
| `modules/comment` | `/comment/*` | Task comments, edit/delete lifecycle with author authorization | `Comment` | Done (tests pass) |
| `modules/activity` | `/activity/*` | Audit trail, state transition logs, history timeline | `Activity` | Done (tests pass) |

### Implemented endpoints and verification

All endpoints are mounted under `/api` in `apps/api/router.go` ahead of the legacy proxy:

1. **`project`** (`/api/project`):
   - `GET /workspace/{workspaceId}`: List projects in workspace.
   - `GET /{id}`: Fetch single project.
   - `POST /{workspaceId}`: Create project.
   - `PUT /{id}`: Update project details and settings.
   - `DELETE /{id}`: Delete project and cascade.
2. **`column`** (`/api/column`):
   - `GET /{projectId}`: List columns for project ordered by position.
   - `POST /{projectId}`: Create column with auto-positioning.
   - `PUT /{id}`: Update column name/icon/color/isFinal.
   - `DELETE /{id}`: Delete empty column.
   - `PUT /reorder/{projectId}`: Batch reorder column positions.
3. **`task`** (`/api/task`):
   - `GET /tasks/{projectId}`: List tasks with filters (`status`, `priority`, `actorId`, `dueBefore`, `dueAfter`, `page`, `limit`, `sortBy`, `sortOrder`).
   - `POST /{projectId}`: Create task (auto-increments number per project, calculates position).
   - `GET /{id}`: Fetch single task with actor profile and attached labels.
   - `PUT /{id}`: Full task update.
   - `DELETE /{id}`: Delete task and cascading dependencies.
   - `PUT /move/{id}`: Move task across projects/statuses.
   - `PUT /status/{id}`, `/priority/{id}`, `/actor/{id}`, `/due-date/{id}`, `/title/{id}`, `/description/{id}`: Single-field updates.
   - `PATCH /bulk`: Batch operations (`updateStatus`, `updatePriority`, `updateActor`, `delete`, `updateDueDate`).
4. **`taskrelation`** (`/api/task-relation`):
   - `GET /{taskId}`: List all relations for a task with target task summary.
   - `POST /`: Create dependency relation (`subtask`, `blocks`, `related`).
   - `DELETE /{id}`: Delete relation.
5. **`label`** (`/api/label`):
   - `GET /{workspaceId}`: List workspace labels.
   - `POST /{workspaceId}`: Create label.
   - `PUT /{id}`: Update label name/color.
   - `DELETE /{id}`: Delete label.
   - `POST /task`: Attach label to task.
   - `DELETE /task`: Detach label from task.
6. **`comment`** (`/api/comment`):
   - `GET /{taskId}`: List task comments ordered chronologically.
   - `POST /{taskId}`: Create comment.
   - `PUT /{id}`: Update comment (author verified).
   - `DELETE /{id}`: Delete comment (author verified).
7. **`activity`** (`/api/activity`):
   - `GET /{taskId}`: List task activities and timeline events.
   - `POST /create`: Create system/custom activity event.
   - `POST /comment`, `PUT /comment`, `DELETE /comment`: Legacy-compatible comment activity endpoints.

All routes are registered in `referenceConfig()` in `apps/api/router.go` for the `/docs` OpenAPI registry. Each module contains a unit test suite running against in-memory SQLite covering normal paths, edge cases, author authorization, and workspace membership enforcement.

## Phase 3: Collaboration breakdown (DONE)

All 5 collaboration modules are fully implemented, tested against in-memory SQLite, mounted ahead of the legacy proxy in `apps/api/router.go`, and verified clean by `filet check .` (116 files, 15,936 lines, 0 errors, 0 warnings).

| Module | Endpoints | Responsibilities | Key schemas | Status |
|---|---|---|---|---|
| `modules/sharelink` | `/share-link/*`, `/shared-project/{token}`, `/share/{token}` | Public read-only shares for projects and tasks | `ShareLink` | Done (tests pass) |
| `modules/notification` | `/notification/*`, `/notifications/*`, `/notification-preferences/*` | In-app notifications, read/unread states, preference toggles | `Notification`, `UserNotificationPreference`, `UserNotificationWorkspaceRule` | Done (tests pass) |
| `modules/timeentry` | `/time-entry/*` | Time tracking, work logs, billable flags | `TimeEntry` | Done (tests pass) |
| `modules/search` | `/search` | Fast multi-domain search across projects, tasks, and comments | Full-text query scoped by membership | Done (tests pass) |
| `modules/workflowrule` | `/workflow-rule/*` | Automated column moves and assignments based on state triggers | `WorkflowRule` | Done (tests pass) |

## Phase 4: Events, realtime, and scheduler (DONE)

All three subsystems are fully implemented, tested against in-memory SQLite and mock/test transports, mounted in `apps/api/main.go` and `apps/api/router.go`, and verified clean by `filet check .` (138 files, 18,912 lines of Go, 0 errors, 0 warnings).

### Subsystems implemented

| Subsystem | Package | Location | Responsibilities & Details | Status |
|---|---|---|---|---|
| **Antenne Event Publisher** | `internal/events` | `apps/api/internal/events/` | Outbound event bus over `antenne-client/go` emitting suite-standard `enveloppe.Event[T]` (`AppOpus`, typed payloads, generated idempotency keys). 10 event channels: `task.{created,updated,deleted,status_changed,actor_changed}`, `project.{created,updated,deleted}`, `comment.created`, `workspace.created`. Asynchronous background dialer with exponential backoff; fail-soft when `ANTENNE_URL` or `ANTENNE_SECRET` is unconfigured. | Done (tests pass) |
| **WebSocket Realtime Hub** | `internal/realtime` | `apps/api/internal/realtime/` | Live WebSocket broadcast at `/api/ws` and `/api/ws/{projectId}` using `gorilla/websocket`. Authenticates via session cookie, bearer header, or `?token=` query param using `porte` session verification. Dual-mode adapter: Redis pub/sub (`REDIS_URL`) with automatic fail-soft in-memory adapter fallback. Connection tracking with `windowId` initiator exclusion. | Done (tests pass) |
| **Maintenance Scheduler** | `internal/scheduler` | `apps/api/internal/scheduler/` | Periodic runner executing background cron maintenance. Prunes expired pending invitations (`expires_at < now`). Evaluates 3 task due-date reminder windows (`one_day_before`, `one_hour_before`, `overdue`), excludes final/completed columns, deduplicates alerts via `task_reminder_sents`, and dispatches in-app alerts to `notifications`. | Done (tests pass) |

All routes and background services are wired into `apps/api/main.go` and `apps/api/router.go` with graceful shutdown lifecycle hooks.

## Dropped modules and integrations (YAGNI)

In accordance with Facile conventions ("features solve real problems, nothing more"), the
following legacy surfaces are dropped and will not be ported:
- **Third-party integrations**: GitHub, Gitea, Slack, Discord, Telegram, and OAuth identity plugins.
- **Direct webhooks**: `workspace-webhook` and `generic-webhook` (Antenne handles event routing).
- **Frontend cleanup**: Delete integration and webhook UI components when decommissioning legacy API.

## Phase 6: Cutover and decommission

1. Implement native S3 / Casier / local file storage for task attachments and assets.
2. Generate baseline goose migration from final schema dump (`pg_dump`).
3. Turn off Drizzle auto-migrate and disable legacy TS API container.
4. Delete `apps/api-legacy/`.
5. Deploy single Go container running the embedded SPA and API.

## Reference implementations

- **Plume**: `apps/api/modules/{auth,clients,verify,documents}` for the composer router pattern.
- **Vision**: `apps/api/modules/{auth,workspaces,sites}` for workspace scoping and clean controllers.
- **Sablier**: `apps/api/main.go` and `internal/env/` for tronc + porte + journal integration.
- **Conventions**:
  - `~/.mycelium/memory/conventions/go-module-architecture.md`
  - `~/.mycelium/memory/conventions/facile-spaces.md`
  - `~/.mycelium/memory/conventions/porte-password-contract.md`

## Exit criteria (per step)

- `go build ./...` compiles cleanly.
- `filet check .` exits 0 with **0 errors and 0 warnings** (never adjust thresholds).
- `go test ./...` and `go vet ./...` pass.
- Contract parity verified against legacy OpenAPI specification.
- Cookie auth + `X-Facile-CSRF` header enforced on all mutating requests.
- Single container successfully serves SPA at `/client` and API at `/api`.

## Gotchas

- Commit with `git commit --no-verify` (lefthook runs biome + full pnpm build).
- `mise run test` fails with older mise versions due to Taskfile depends syntax; use `go test ./...` directly.
- Always run Go commands from `apps/api/` or reference the module path.
- Keep `golang:1.26-alpine` in Dockerfile and `go 1.26` in `go.mod` in sync.
- Never use inline comments or comments in function bodies in Go code.