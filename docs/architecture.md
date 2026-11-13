# Opus: architecture

How a request travels through Opus, what the database holds, how sign-in works, and how Opus
exchanges events with the rest of the suite.

## Runtime topology

Opus ships as one container running one Go binary, `opus-api`. It serves the API, the MCP
server, the WebSocket hub and the built React client on one port.

```
Internet ──▶ Traefik ──▶ :80 opus-api ──┬──▶ /health, /ready, /version, /docs
                                         ├──▶ /mcp            MCP server (with OIDC and MCP_RESOURCE)
                                         ├──▶ /ws, /api/ws    realtime WebSocket hub
                                         ├──▶ /api/*          the API
                                         └──▶ everything else /client, SPA fallback to index.html
                                                   │
                        ┌──────────────┬───────────┼───────────┬──────────────┐
                   Postgres 16     Registre     Antenne     Redis (opt)    S3 (opt)
```

`deploy/opus-entrypoint.sh` is the container command. It derives `OPUS_API_URL` and
`DATABASE_URL` when they are unset, runs `apps/web/env.sh` to write the real URLs into the built
JavaScript, then `exec`s the binary. See [deployment.md](deployment.md).

## Components

| Component | Path | Role |
|---|---|---|
| API | `apps/api` | Go on the tronc chassis: chi router, GORM over Postgres, goose migrations. One package per domain under `modules/`, shared plumbing under `internal/` |
| Web client | `apps/web` | React 19 SPA: TanStack Router file routes, TanStack Query for server state, Zustand for client state |
| Shared libs | `packages/libs` | TypeScript helpers for the client |
| Translations | `i18n/` | JSON per locale, checked by `scripts/i18n/` |

## Boot

`main.go` runs, in order:

1. Load the configuration (`internal/env`). A bad value stops the boot.
2. Create the database when it is missing, open it, and apply the goose migrations under a
   Postgres session lock. A failed migration stops the boot with exit code 1.
3. Build the porte auth kits: sessions, the user store (`kit.Store`), `porte/oidc` and the
   Registre deletion receiver.
4. Start the realtime hub, the Antenne publisher and outbox drainer, the inbound Sablier sync, the
   scheduler and the Registre space sync.
5. Build the router and serve until SIGINT or SIGTERM, then shut down gracefully.

## Request lifecycle

1. `tronc/httpx` builds the router: real client IP from trusted proxies, request logging, CORS
   (`CORS_ORIGINS`, credentials allowed, plus the `X-Facile-CSRF` and `X-Opus-Window-Id`
   headers), and security headers with a Content-Security-Policy that hashes the client's inline
   scripts.
2. chi matches the path. Root routes come first, then `/api`, then the SPA catch-all, which is
   mounted last.
3. Under `/api`, `realtime.InitiatorMiddleware` reads `X-Opus-Window-Id`, so a broadcast skips the
   browser tab that caused the change.
4. Protected routes run `middleware.RequireAuth`, which resolves a porte session from the
   `__Host-session` cookie or an `Authorization: Bearer` token (a personal API token is a
   labelled session), or a Registre machine JWT when `OIDC_MACHINE_AUDIENCE` is set. A
   cookie-authenticated request that is not GET, HEAD or OPTIONS must also send
   `X-Facile-CSRF`.
5. The module controller checks workspace membership and calls its service, which writes through
   GORM, publishes to the realtime hub, and queues Antenne events in the same transaction.
6. Errors are tronc `errors` values, written as JSON by `httpjson.WriteError`.

## Data model

PostgreSQL 16. The schema is the SQL in `apps/api/migrations/`; `apps/api/schemas/` holds the GORM
structs that mirror it. Ids are `int64`.

- **Identity**: `users`, and porte's `porte_identities`, `porte_sessions` and
  `porte_login_codes`. `sessions`, `accounts` and `verifications` are leftovers of the retired
  Better Auth schema, kept by the baseline.
- **Tenancy**: `workspaces`, `workspace_members`, `invitations`. A workspace is linked to a
  Registre space when `registre_space_id` is set, and is someone's personal workspace when
  `personal_owner_id` is set.
- **Work**: `projects` own `columns`, `workflow_rules` and `share_links`. `tasks` belong to a
  project and a column and carry `activities`, `time_entries`, `task_relations`, `task_labels`,
  `asset` rows and `task_reminder_sents`. `labels` are workspace-scoped. Comments are
  `activities` rows of type `comment`; the `comments` table is the older surface.
- **Notifications**: `notifications`, `user_notification_preferences` and
  `user_notification_workspace_rules`.
- **GitHub**: `project_github_repos` and `task_pull_requests`.
- **Suite plumbing**: `antenne_outbox` (the library's outbox), `pool_processed_events` (inbound
  idempotency), `mcp_actions` (pending MCP confirmations), `forfait_state` (the last plan
  document from Registre) and `directory_deletion_state` (the deletion cursor).

Projects and tasks carry a `facile_id`, the identity they share with Sablier.

## Authentication

Opus signs people in through Registre only, with porte. It has no password, sign-up or magic-link
route.

- **Browser**: `GET /api/auth/oidc` starts an OIDC code flow with PKCE, and
  `/api/auth/oidc/callback` finishes it, upserts the user (`kit.Store`), sets the session cookie
  and redirects to `OIDC_SUCCESS_URL`. Profile and picture follow Registre at every sign-in.
- **CLI**: with `OIDC_CLI_AUDIENCE` set, `POST /api/auth/oidc/exchange` trades a device-grant
  token for a session. Personal API tokens are minted at `/api/auth/api-tokens`.
- **Other suite apps**: with `OIDC_MACHINE_AUDIENCE` and `OIDC_SCOPE_PREFIX=opus`, a Registre
  machine JWT needs `opus.read` for reads and `opus.write` for writes.
- **MCP clients**: `porte/mcpauth` checks a Registre token bound to `MCP_RESOURCE`.

`porte/directory` keeps Registre's shared spaces in step with Opus workspaces: linked workspaces
answer 409 to rename, delete, invitation and role changes. The same package receives Registre's
deletion notices, by poll and by signed push on `POST /api/registre/deletions`, and
`kit.EraseAccount` erases the person in one transaction. A subject Registre has erased never
gets an account again. See [erasure.md](erasure.md).

## Real-time

`/ws/{projectId}` (also under `/api`) upgrades to a WebSocket after authenticating the session
cookie, a bearer token or a `?token=` query parameter, and checking workspace membership.
Connections are grouped per project. Broadcasts go through Redis pub/sub when `REDIS_URL` is set
and stay in memory otherwise, which is only correct for one container.

The scheduler runs every five minutes: it prunes expired invitations and sends due-date
reminders as in-app notifications, one day before, one hour before and when overdue.

## Antenne bus

Opus syncs projects and tasks with Sablier over Antenne, and mirrors Sablier's time entries.
`internal/events` publishes, `internal/poolsync` applies inbound events.

```
Opus write ──▶ antenne_outbox (same transaction) ──▶ drainer ──▶ Antenne
     ▲                                                              │
     └──── poolsync ◀── pool_processed_events (idempotency) ◀───────┘
```

- Events go into `antenne_outbox` first. The drainer sends unsent rows in id order and stops at
  the first failure, so ordering holds. Sent rows are kept 7 days.
- Every boot's reconcile queues `project.created` and `task.created` again for every syncable
  project, so a receiver that missed events catches up.
- Payloads carry the workspace's space (Registre space, or personal plus the owner), and projects
  of a shared workspace that is not linked to Registre are never sent.
- Opus also listens to `git.pull_request` and `git.push`, which Antenne fans out from GitHub, and
  handles them like the direct webhook.

The wire contract is in [sablier-sync.md](sablier-sync.md). `nook.yaml` at the repo root lists
the six project and task channels for reference.

## Object storage

Task images go to S3-compatible storage through presigned URLs when S3 is configured
(`PUT /api/task/image-upload/{id}`, then `POST /api/task/image-upload/{id}/finalize`).
Without S3 the API takes the upload itself and writes it under `UPLOAD_DIR/assets`. Downloads go
through `GET /api/asset/{id}`, which checks workspace membership, or an active share link with
`?share_token=` when the project is not public. Workspace logos always sit on the local disk under
`UPLOAD_DIR/workspaces`, which the compose file backs with the `opus-uploads` volume.

## MCP and Ideas

The MCP server at `/mcp` (Streamable HTTP, stateless) exposes eleven tools over projects and
tasks. Tools call the same controllers as the HTTP routes, through `Calls()`, so permissions,
broadcasts and Antenne events match. Deletes wait for the person's confirmation, by elicitation
or on the `/pending` page. The tools are listed in the [README](../README.md#mcp).

With `BULLE_URL` set, `/api/ideas` proxies Bulle's Ideas for the active workspace, behind the
plan entitlements stored in `forfait_state`.
