# Phase 6 Handoff: Cutover and Decommission

Cold-start handoff for a fresh session picking up Opus backend development.

## Current State

- **Branch**: `feat/go-backend`
- **Head commit**: `ac3d3451`
- **Completed phases**:
  - **Phase 0**: Scaffold + dual-container strangler setup.
  - **Phase 1 (K1 + K2)**: Go auth seam (`porte` over `portepg`, `/auth/me`, `/auth/register`, `/auth/login`, `/auth/change-password`), workspace CRUD, invitations lifecycle, legacy proxy HMAC auth bridge (`AUTH_BRIDGE_SECRET`).
  - **Phase 1.5**: Baseline filet refactor; composer pattern applied to all routers; 0 warnings baseline achieved.
  - **Phase 2**: Core domain modules (`project`, `column`, `task`, `taskrelation`, `label`, `comment`, `activity`) fully Go-native with in-memory SQLite unit tests.
  - **Phase 3**: Collaboration modules (`sharelink`, `timeentry`, `workflowrule`, `notification`, `search`) fully Go-native with table-driven unit tests.
  - **Phase 4**: Events, realtime pub/sub, and scheduler:
    - `internal/events`: Antenne outbound event publisher with standard `enveloppe.Event[T]` structure across 10 channels (`task.*`, `project.*`, `comment.created`, `workspace.created`), background connection with exponential backoff, and fail-soft fallback.
    - `internal/realtime`: Live WebSocket broadcast hub at `/api/ws` and `/api/ws/{projectId}`, cookie/bearer/token auth via `porte` session verification, `windowId` initiator exclusion, and dual-mode adapter (Redis pub/sub via `REDIS_URL` with fail-soft in-memory fallback).
    - `internal/scheduler`: Background maintenance runner for stale invitation pruning (`expires_at < now`) and 3-window due-date reminder alerts (`one_day_before`, `one_hour_before`, `overdue`), deduplicated via `task_reminder_sents`.
    - Wired into `apps/api/main.go` and `apps/api/router.go` with graceful shutdown.
- **Quality verification**:
  - `filet check .` outputs 0 errors and 0 warnings across 138 files (18,912 lines of Go).
  - `go test -count=1 ./...` passes across all 20 packages in `apps/api`.
  - `go build ./...` compiles cleanly.

---

## What to Do Next: Phase 6 (Cutover and Decommission)

Phase 5 (third-party plugins like Slack, Discord, Telegram, GitHub and direct webhooks) was intentionally dropped under Facile YAGNI conventions (Antenne serves as the sole outbound event bus).

The next task is **Phase 6: Cutover and Decommission**:

### 1. Native Asset & File Storage
- Port task attachment and file upload handling to Go.
- Support local file storage and S3 storage compatible with the suite standards.
- Replaces legacy asset handling in `apps/api-legacy/src/storage`.

### 2. Database Migration Baseline
- Dump live database schema using `pg_dump`.
- Generate a clean baseline `goose` migration in `apps/api/migrations/`.
- Replace GORM `AutoMigrate` schema ownership with versioned goose migrations.

### 3. Decommission Legacy Backend
- Turn off Drizzle auto-migrate and disable the legacy TypeScript API container.
- Remove the legacy proxy route from `apps/api/router.go`.
- Delete `apps/api-legacy/`.

### 4. Frontend UI Cleanup
- Remove dead UI surfaces in `apps/web`:
  - Third-party integration settings (GitHub, Gitea, Slack, Discord, Telegram).
  - Custom webhook endpoints (replaced by Antenne).
  - Deprecated plugins.

### 5. Single-Container Dokploy Deployment
- Update `Dockerfile` to build the Go binary and embed the built SPA from `apps/web/dist` at `/client`.
- Verify single container serves both SPA (`/client`) and API (`/api`).

---

## Architectural Rules and Constraints

1. **5-file module architecture**:
   - Every module package under `apps/api/modules/<feature>/` strictly uses:
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
   - Documentation: Top-level exported types and functions must have doc comments (`requireDocComments: true`).
3. **Workspace multi-tenancy**:
   - Workspace ID is a strict security boundary, not an optional filter.
   - Always verify caller membership in `workspace_members` before reading or mutating workspace data.
4. **Git commits**:
   - Commit with `git commit --no-verify` (lefthook runs full pnpm build which has legacy frontend errors).
   - Conventional Commits: `type(scope): summary`, imperative, lowercase, under 72 chars.

---

## Running Verification Checks

```bash
# Run unit tests across all packages
cd apps/api && go test -count=1 ./...

# Verify Go compilation
cd apps/api && go build ./...

# Run filet code quality gate (from repo root)
filet check .
```
