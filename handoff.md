# Phase 4 Handoff: Events, Realtime, and Scheduler

## Current State

- **Branch**: `feat/go-backend`
- **Completed phases**:
  - **Phase 0**: Scaffold + dual-container strangler setup.
  - **Phase 1 (K1 + K2)**: Go auth seam (porte over portepg, `/auth/me`, `/auth/register`, `/auth/login`, `/auth/change-password`), workspace CRUD, invitations lifecycle, legacy proxy HMAC auth bridge (`AUTH_BRIDGE_SECRET`).
  - **Phase 1.5**: Baseline filet refactor; composer pattern applied to all routers.
  - **Phase 2**: Core domain modules (`project`, `column`, `task`, `taskrelation`, `label`, `comment`, `activity`) fully Go-native with in-memory SQLite unit tests.
  - **Phase 3**: Collaboration modules (`sharelink`, `timeentry`, `workflowrule`, `notification`, `search`) fully Go-native with table-driven unit tests.
- **Quality verification**:
  - `filet check .` outputs `clean 117 files, 16211 lines, nothing to complain about` (0 errors, 0 warnings).
  - `go test -count=1 ./...` passes across all packages.
  - `go build ./...` compiles cleanly.

---

## Phase 4 Objectives

Phase 4 ports events, realtime pub/sub, and background job scheduling into the Go backend.

### 1. Antenne Event Publisher

- **Protocol**: Suite standard outbound event bus via `antenne` client.
- **Events to publish**:
  - `task.created`, `task.updated`, `task.deleted`, `task.status_changed`, `task.actor_changed`
  - `project.created`, `project.updated`, `project.deleted`
  - `comment.created`
  - `workspace.created`
- **Envelope contract**:
  - Envelopes follow the Facile `enveloppe` structure, keyed on `actor_email` or `user_id`.
- **Note on dropped legacy code**:
  - `workspace-webhook` and `generic-webhook-integration` are intentionally dropped. Antenne handles downstream webhook dispatching and event subscriptions.

### 2. WebSocket Broadcast (`/api/ws`)

- **Route**: `GET /api/ws` (Gorilla WebSocket or similar standard Go websocket handler).
- **Authentication**: Authenticate caller via session cookie / bearer token on connection handshake.
- **Subscriptions**: Workspace-scoped channels (`workspace:{workspaceId}`).
- **Backing bus**: Redis pub/sub (using `REDIS_URL` with fail-soft in-memory channel broadcast when Redis is unreachable or unconfigured in local dev).
- **Payloads**: Realtime events to notify web clients of live task updates and column position changes.

### 3. Background Maintenance Scheduler

- **Mechanism**: Go ticker / runner executing periodic cron jobs.
- **Tasks**:
  - Stale invitation cleanup (pruning expired invitations).
  - Due-date notification generator (finding tasks with `due_date <= now` and dispatching in-app notifications).
  - Reminder delivery.

---

## Architectural Rules and Constraints

1. **5-file architecture**:
   - Every module package under `apps/api/modules/<feature>/` must strictly use:
     - `types.go`: DTO request and response structs only.
     - `router.go`: Chi route declarations with composer pattern. No database queries.
     - `controller.go`: Request validation, mapping, error handling.
     - `service.go`: Business logic and GORM DB operations.
     - `documentation.go`: OpenAPI route registry (`Documentation() documentation.Module`).
     - `service_test.go`: Unit tests against SQLite in-memory (`:memory:`).
2. **Filet limits**:
   - File length: Maximum 300 lines per file (`limits.fileLines: 300`).
   - Function length: Maximum 60 lines (`limits.funcLines: 60`).
   - Statements: Maximum 40 statements per function (`limits.funcStatements: 40`).
   - Cognitive complexity: Maximum 15 (`limits.complexity: 15`).
   - Return values: Maximum 3 (`limits.returns: 3`).
   - Struct fields: Maximum 15 (`limits.structFields: 15`). Embed sub-structs if larger.
   - Code style bans: NEVER add inline comments. NEVER add comments inside function bodies.
3. **Workspace multi-tenancy**:
   - Always verify that the caller is a member in `workspace_members` before reading or mutating workspace data.
4. **Git commits**:
   - Commit with `git commit --no-verify` (lefthook runs full pnpm build which has legacy frontend errors).
   - Conventional Commits: `type(scope): summary`, imperative, lowercase, under 72 chars.

---

## Running Checks

```bash
# Run unit tests across all packages
cd apps/api && go test -count=1 ./...

# Run filet code quality gate
filet check .

# Verify Go compilation
cd apps/api && go build ./...
```
