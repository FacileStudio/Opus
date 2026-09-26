# Phase 6 Handoff: Cutover and Decommission

## Current State

- **Branch**: `feat/go-backend`
- **Completed phases**:
  - **Phase 0**: Scaffold + dual-container strangler setup.
  - **Phase 1 (K1 + K2)**: Go auth seam (porte over portepg, `/auth/me`, `/auth/register`, `/auth/login`, `/auth/change-password`), workspace CRUD, invitations lifecycle, legacy proxy HMAC auth bridge (`AUTH_BRIDGE_SECRET`).
  - **Phase 1.5**: Baseline filet refactor; composer pattern applied to all routers.
  - **Phase 2**: Core domain modules (`project`, `column`, `task`, `taskrelation`, `label`, `comment`, `activity`) fully Go-native with in-memory SQLite unit tests.
  - **Phase 3**: Collaboration modules (`sharelink`, `timeentry`, `workflowrule`, `notification`, `search`) fully Go-native with table-driven unit tests.
  - **Phase 4**: Events, realtime, and scheduler:
    - `internal/events`: Antenne event publisher with `enveloppe.Event` standard structure.
    - `internal/realtime`: WebSocket live broadcast hub at `/api/ws` and `/api/ws/{projectId}`, in-memory and Redis pub/sub adapters, session cookie/token authentication, windowId initiator exclusion.
    - `internal/scheduler`: Background maintenance runner for stale invitation pruning and due date notification generation.
- **Quality verification**:
  - `filet check .` outputs 0 errors and 0 warnings across 138 files (18,912 lines).
  - `go test -count=1 ./...` passes across all packages.
  - `go build ./...` compiles cleanly.

---

## Phase 6 Objectives: Cutover and Decommission

Phase 5 (third-party plugins and direct webhooks) was dropped under Facile YAGNI conventions. Phase 6 is the final decommissioning step:

1. **Native File & Asset Storage**:
   - Port task attachment and asset upload handling to Go (S3/local storage).
2. **Database Migration Baseline**:
   - Generate baseline goose migration from final schema dump (`pg_dump`).
3. **Decommission Legacy Backend**:
   - Stop Drizzle migrations and disable the legacy TypeScript API container.
   - Delete `apps/api-legacy/`.
4. **Frontend Cleanup**:
   - Delete legacy webhook and integration UI components in `apps/web`.
5. **Single-Container Dokploy Deployment**:
   - Deploy single container running Go API and embedded React SPA at `/client`.

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
