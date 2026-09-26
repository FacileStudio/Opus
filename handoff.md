# Phase 6 Handoff: Cutover and Decommission (DONE)

Cold-start handoff for a fresh session picking up Opus backend development.

## Current State

- **Branch**: `feat/go-backend`
- **Head status**: All phases (0, 1, 1.5, 2, 3, 4, 6) complete.
- **Port complete**:
  - **Phase 0**: Scaffold + dual-container strangler setup.
  - **Phase 1 (K1 + K2)**: Go auth seam (`porte` over `portepg`, `/auth/me`, `/auth/register`, `/auth/login`, `/auth/change-password`), workspace CRUD, invitations lifecycle, legacy proxy HMAC auth bridge (`AUTH_BRIDGE_SECRET`).
  - **Phase 1.5**: Baseline filet refactor; composer pattern applied to all routers; 0 warnings baseline achieved.
  - **Phase 2**: Core domain modules (`project`, `column`, `task`, `taskrelation`, `label`, `comment`, `activity`) fully Go-native with in-memory SQLite unit tests.
  - **Phase 3**: Collaboration modules (`sharelink`, `timeentry`, `workflowrule`, `notification`, `search`) fully Go-native with table-driven unit tests.
  - **Phase 4**: Events, realtime pub/sub, and scheduler:
    - `internal/events`: Antenne outbound event publisher with standard `enveloppe.Event[T]` structure across 10 channels.
    - `internal/realtime`: Live WebSocket broadcast hub at `/api/ws` and `/api/ws/{projectId}`, cookie/bearer/token auth via `porte` session verification, `windowId` initiator exclusion, dual-mode adapter (Redis pub/sub via `REDIS_URL` with fail-soft in-memory fallback).
    - `internal/scheduler`: Background maintenance runner for stale invitation pruning and 3-window due-date reminder alerts.
  - **Phase 6**: Cutover and decommission:
    - `apps/api/modules/asset`: Strict 5-file architecture (`types.go`, `router.go`, `controller.go`, `service.go`, `documentation.go`, `service_test.go`).
    - `internal/storage`: Enhanced with `asset.go` and `s3.go` supporting local disk and S3 with AWS SigV4 presigned URLs.
    - `apps/api/migrations`: `00001_baseline_opus.sql` declaring all production tables and indexes under Goose standard.
    - `internal/database`: `RunMigrationsContext` using `pressly/goose/v3` Provider API with Postgres session locking and GORM AutoMigrate fallback.
    - `apps/api/router.go`: Legacy proxy removed. Backend is 100% Go-native.
    - `apps/web`: Removed dropped third-party integration settings (GitHub, Gitea, Slack, Discord, Telegram) and custom webhook UI surfaces; verified clean build (`pnpm --filter @opus/web build`).
    - `Dockerfile` & `deploy/opus-entrypoint.sh`: Updated for single-container Dokploy deployment running only `/usr/local/bin/opus-api` and serving `/client` and `/api`.
- **Quality verification**:
  - `filet check .` outputs **0 errors, 0 warnings** across 147 files (19,946 lines of Go).
  - `go test -count=1 ./...` passes across all 21 packages in `apps/api`.
  - `go build ./...` compiles cleanly.
  - `pnpm --filter @opus/web build` succeeds in ~31s.

---

## Running Verification Checks

```bash
# Run unit tests across all packages
cd apps/api && go test -count=1 ./...

# Verify Go compilation
cd apps/api && go build ./...

# Run filet code quality gate (from repo root)
filet check .

# Verify web build
pnpm --filter @opus/web build
```
