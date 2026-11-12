# Opus

Self-hosted project management for the Facile Suite: workspaces, projects, boards, tasks,
comments, time entries and share links. A Go API serves the data and the built React client
from one binary. Sign-in goes through Registre only. Project and task changes sync with
Sablier over the Antenne bus.

Simplicity first. Do not add features, refactors or improvements beyond what was asked.

## Tech stack

- **API** (`apps/api`): Go 1.26.11, chi v5.5.0, GORM over PostgreSQL 16, ordered goose SQL
  migrations embedded in the binary
- **Auth**: [`porte`](https://github.com/FacileStudio/porte) v0.19.0. `porte/kit` owns the user
  store and erasure, `porte/oidc` signs people in through Registre, `porte/directory` v0.5.0
  syncs Registre spaces and deletion notices, `porte/mcpauth` v0.2.0 guards `/mcp`,
  `porte/forfait` v0.1.1 holds plan entitlements
- **Chassis**: [`tronc`](https://github.com/FacileStudio/tronc) v0.22.0 supplies `env`, `httpx`,
  `middleware`, `logger`, `health`, `healthcheck`, `spa`, `apiref` and `mcp`;
  `tronc/testdb` v0.1.1 gives tests a real Postgres
- **Events**: [`antenne-client/go`](https://github.com/FacileStudio/antenne-client) v1.1.0, with
  its durable outbox (`outbox`), envelope types (`envelope`) and channel catalog (`catalog`)
- **Client** (`apps/web`): React 19, Vite 7, TanStack Router (file-based) and Query, Tailwind
  CSS 4, Zustand, base-ui and Radix, Tiptap, react-i18next. It does not use muse
- **Shared**: `packages/libs` (TypeScript helpers), `i18n/` (translation JSON)
- **Tooling**: pnpm 10.32.1 and Turbo for the JavaScript side, mise for Go and Node, lefthook
  for git hooks, Biome for lint and format, Vitest for client tests

## Key commands

All commands run from the repo root unless noted.

```sh
mise install                              # Node 22 and lefthook; postinstall runs lefthook install
pnpm install --frozen-lockfile            # JavaScript dependencies
pnpm --filter @opus/web dev               # Vite dev server on :5173
pnpm --filter @opus/web build             # build the client into apps/web/dist
pnpm --filter @opus/web typecheck         # tsc --noEmit
pnpm --filter @opus/web exec biome check .  # client lint, as CI runs it
pnpm --filter @opus/web test              # Vitest
pnpm i18n:check                           # validate the translation files
filet check                               # Go rules from filet.yml
filet gate .                              # the pre-push gate: gofmt, rules, vet, staticcheck, tests
```

### API

Run these from `apps/api`.

```sh
go run .                                  # the API; reads .env in the working directory
go vet ./...
go test ./...                             # database tests skip without TEST_DATABASE_URL
douane scan -fail critical -scope prod .  # advisory scan, as CI and pre-push run it
mise run dev                              # the API on :1337, where the client's dev build points
```

Database tests need a throwaway Postgres. `TEST_DATABASE_URL` is a full URL with a database
name; each test package gets its own schema:

```sh
docker run -d --name opus-test -e POSTGRES_PASSWORD=opus -p 127.0.0.1:5433:5432 \
  postgres:16 -c fsync=off -c synchronous_commit=off -c full_page_writes=off
TEST_DATABASE_URL='postgres://postgres:opus@127.0.0.1:5433/postgres?sslmode=disable' go test ./...
```

`fsync=off` matters: every test truncates its tables, which costs seconds per test with fsync on.

## Deployment shape

Production is a Dokploy compose on `opus.facile.studio`. A push to `main` deploys it. A `v*` tag
only builds the GHCR image and the GitHub release; it deploys nothing.

`docker-compose.yml` runs two services:

- `opus-db`: Postgres 16. `deploy/db/init/01-opus-go.sql` creates the `opus_go` database,
  which holds production data. The `opus` database in the same container is the retired
  TypeScript app's.
- `opus-app`: the one Go binary, built by the root `Dockerfile`.

The image has three stages. pnpm builds the client, Go builds the API with the version from
`package.json` (`scripts/version.sh`, set through `-ldflags -X main.version`), and both land in
`alpine:3.21` as user `appuser`: binary at `/usr/local/bin/opus-api`, client at `/client`,
`PORT=80`. The health check re-execs the binary (`opus-api healthcheck`, from
`tronc/healthcheck`).

`deploy/opus-entrypoint.sh` runs first. It derives `DATABASE_URL` from `POSTGRES_*` when unset
(database `opus_go`, host `postgres`), lets `GO_DATABASE_URL` override it, derives
`OPUS_API_URL` as `<OPUS_CLIENT_URL>/api`, then runs `apps/web/env.sh`, which rewrites the
placeholder strings `OPUS_API_URL` and `OPUS_CLIENT_URL` inside the built JavaScript.

Routing inside the binary (`apps/api/router.go`):

```
/health, /ready, /version, /docs   -> root routes
/mcp                               -> the MCP server, only with OIDC_ISSUER and MCP_RESOURCE
/ws, /ws/{projectId}               -> realtime WebSocket (also under /api)
/api/*                             -> the API
everything else                    -> the SPA, mounted last with tronc/spa
```

Opus has no `/api/version`; a 404 there is normal. After a deploy, Traefik answers
`404 page not found` on every path for about 30 seconds while the new container boots.

`charts/opus` is a Helm chart. Production does not use it.

## Project structure

```
.
├── apps/
│   ├── api/                  # Go API (module github.com/FacileStudio/Opus/apps/api)
│   │   ├── internal/
│   │   │   ├── authcontext/  # Request-scoped identity, including the MCP actor
│   │   │   ├── database/     # GORM connection, EnsureDatabase, goose runner, dbtest/
│   │   │   ├── documentation/# Types the API reference registry is built from
│   │   │   ├── env/          # Config loader on top of tronc/env
│   │   │   ├── events/       # Antenne publisher, outbox drain, payloads, reconcile
│   │   │   ├── middleware/   # RequireAuth and OptionalAuth over a porte session
│   │   │   ├── poolsync/     # Applies inbound Sablier project and task events
│   │   │   ├── realtime/     # WebSocket hub, optional Redis fan-out
│   │   │   ├── scheduler/    # Invitation pruning and due-date reminders
│   │   │   ├── storage/      # Local uploads and S3 assets
│   │   │   └── wireid/       # IDs that decode from a JSON number or a decimal string
│   │   ├── modules/          # One package per domain: auth, workspace, project, task, mcp, ...
│   │   ├── migrations/       # Ordered goose SQL, embedded
│   │   ├── schemas/          # GORM structs mirroring the SQL
│   │   ├── main.go           # Boot: config, database, auth, services, server
│   │   └── router.go         # buildRouter: root routes, /api, then the SPA
│   └── web/                  # React client (@opus/web)
├── packages/libs/            # Shared TypeScript helpers
├── i18n/                     # Translations; scripts/i18n/ validates them
├── deploy/                   # Container entrypoint and Postgres init scripts
├── docs/                     # Architecture, erasure, Sablier sync, API notes
├── scripts/version.sh        # Prints v<package.json version>
├── Dockerfile                # The only Dockerfile
└── docker-compose.yml
```

`tests/` holds TypeScript tests for the retired Hono API. Nothing runs them.

## Data model

Workspaces hold projects, projects hold columns and tasks. A workspace is linked to a Registre
space when `registre_space_id` is set, and is someone's personal workspace when
`personal_owner_id` is set. Comments are `activities` rows of type `comment`, not rows in the
`comments` table. `users` and the `porte_*` tables are identity.

`antenne_outbox` is the library's outbox table (migration 00009). `mcp_actions` holds pending
MCP confirmations, `forfait_state` the last plan document, and `directory_deletion_state` the
Registre deletion cursor.

## Environment variables

The API reads `.env` from its working directory (godotenv), then the process environment.

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | none | Postgres connection string. The entrypoint builds it from `POSTGRES_*` when unset |
| `POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_HOST` / `POSTGRES_PORT` | `opus_go` / `opus` / none / `postgres` / `5432` | Entrypoint only. It exits 1 when neither `DATABASE_URL` nor `POSTGRES_PASSWORD` is set |
| `GO_DATABASE_URL` | empty | Entrypoint only. Replaces `DATABASE_URL` when set |
| `PORT` | `8080`, `80` in the image | Listen port. `mise run dev` uses 1337 |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error`. Anything else stops the boot |
| `CORS_ALLOWED_ORIGINS` | empty | tronc reads the first set of `CORS_ALLOWED_ORIGINS`, `ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`, `TRUSTED_ORIGINS`, `CLIENT_ORIGIN`. Production sets `CORS_ORIGINS` |
| `TRUSTED_PROXIES` | empty | Proxies whose forwarded headers the router trusts |
| `JOURNAL_URL` / `JOURNAL_TOKEN` | empty | Log shipping to Journal when both are set |
| `CLIENT_DIR` | `./client`, `/client` in the image | Where the built SPA lives. No client, no SPA mount |
| `UPLOAD_DIR` | `./data/uploads`, `/app/data/uploads` from the entrypoint | Workspace logos and local assets |
| `OIDC_ISSUER` | empty | `https://sso.facile.studio`. Empty disables sign-in |
| `OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` / `OIDC_REDIRECT_URL` | empty | Required when `OIDC_ISSUER` is set. Redirect: `https://opus.facile.studio/api/auth/oidc/callback` |
| `OIDC_SUCCESS_URL` | first CORS origin | Where a completed login lands |
| `OIDC_CLI_AUDIENCE` | empty | Enables porte's device exchange for the CLI |
| `OIDC_MACHINE_AUDIENCE` / `OIDC_SCOPE_PREFIX` | empty | Offline verification of Registre machine JWTs; with prefix `opus`, GET needs `opus.read` and writes need `opus.write` |
| `MCP_RESOURCE` | empty | Public MCP URL. With `OIDC_ISSUER` it turns `/mcp` on |
| `PORTE_TOKEN_KEY` | empty | Seals stored OIDC tokens. When set, boot reseals existing rows in the background |
| `ANTENNE_URL` (or `ANTENNE_INSTANCE`) | empty | Antenne instance. Empty turns the bus off |
| `ANTENNE_SECRET` | empty | Required with a URL: a URL without it stops the boot |
| `ANTENNE_INSTANCE_ID` | empty | Distinguishes two Opus deployments on one bus |
| `ANTENNE_SYNC_WORKSPACE_ID` | empty | Fallback workspace for Sablier payloads that carry no space |
| `BULLE_URL` | empty | Bulle base URL for the Ideas page. Needs OIDC too |
| `REDIS_URL` | empty | Realtime fan-out across replicas. Empty keeps it in memory |
| `GITHUB_WEBHOOK_SECRET` | empty | Verifies the GitHub webhook signature |
| `S3_ENDPOINT` / `S3_BUCKET` / `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY` | empty | All four turn S3 assets on. Also `S3_REGION`, `S3_PUBLIC_BASE_URL`, `S3_FORCE_PATH_STYLE`, `S3_MAX_IMAGE_UPLOAD_BYTES`, `S3_PRESIGN_TTL_SECONDS` |
| `OPUS_CLIENT_URL` / `OPUS_API_URL` | empty | Container only. Substituted into the built JavaScript by `env.sh` |

## Conventions

- **Migrations are ordered SQL** in `apps/api/migrations/`, applied at boot by
  `internal/database.RunMigrations` (goose, Postgres session lock). `AutoMigrate` never runs.
  Add a numbered file with `-- +goose Up` and `-- +goose Down`, and update the struct in
  `schemas/` by hand. goose must stay at v3.27.3 or later: older versions look for their
  ledger in `public` and fail in the per-schema test databases.
- **Postgres only.** CI fails on a SQLite, MySQL, SQL Server or ClickHouse driver in `go.mod`.
  Tests open a database through `internal/database/dbtest`.
- **Routes register inside `router.Route("/api", ...)`.** Only `/health`, `/ready`, `/version`,
  `/docs`, `/mcp` and `/ws` live at the root. The SPA catch-all is registered last.
- **Document every route.** Each module has a `documentation.go`; `referenceConfig()` in
  `router.go` collects them for `/docs`, and `router_test.go` fails when a module is missing.
- **New in-process callers go through the module controllers' `Calls()`**, as the MCP server
  does (`newDomain` in `router.go`). The services' raw methods skip realtime and Antenne events.
- **Request IDs use `internal/wireid`.** The client sends IDs as strings and tronc decodes
  strictly, so an `int64` field 400s.
- **Events go through the outbox.** `internal/events` enqueues in a transaction with
  `outbox.Enqueue`; the drainer sends rows in id order and stops at the first failure.
- **Client package manager is pnpm.** `bun install` rewrites the root `package.json` workspaces.
  The Docker client stage runs `pnpm install --frozen-lockfile`.
- **Commits**: Conventional Commits. The shared commit-msg hook (`FacileStudio/hooks`) adds the
  gitmoji. Pre-commit runs `pnpm exec biome ci .` and `pnpm run build`; pre-push runs
  `filet gate .` and douane on `apps/api`.
- **Releases**: a `## [x.y.z] - YYYY-MM-DD` section in `CHANGELOG.md`, the same version in
  `package.json`, an annotated `vx.y.z` tag. `release.yml` refuses a tag whose version or
  changelog section does not match.

## Gotchas

- **The outbox re-encodes every row.** Postgres `jsonb` reorders keys, so `canonical()` in
  `internal/events/outbox.go` decodes each row into its payload type and marshals it again.
  A new payload type needs a case there, or receivers get Postgres's key order. v2.35.2 fixed a
  regression of exactly this.
- **Sent outbox rows stay 7 days** (the library's default retention), and every boot's
  reconcile queues `project.created` and `task.created` for every syncable project again.
  Unsent rows are never pruned.
- **Erasure runs in one transaction.** `kit.EraseAccount` calls `workspace.Eraser.EraseData`,
  which deletes stored objects, invitations to the address, pending MCP actions and every
  outbox row naming the person's email or subject, sent or queued. `docs/erasure.md` is the
  rule; `modules/workspace/erasure_test.go` scans every text column for leftovers.
- **Auth is porte through Registre only.** There is no password or sign-up route, so without
  `OIDC_ISSUER` nobody can sign in. Cookie-authenticated mutations need the `X-Facile-CSRF`
  header; bearer requests do not.
- **The client does not call `/api` relatively.** It builds URLs from `VITE_API_URL`, which the
  production build sets to the placeholder `OPUS_API_URL` and `env.sh` rewrites at boot. Build
  every URL with `apps/web/src/fetchers/get-api-url.ts`; hand-built bases produced `/api/api`.
- **Linked workspaces are owned by Registre.** Rename, delete, invite and role changes on a
  workspace with `registre_space_id` answer 409.
- **douane cannot read `pnpm-lock.yaml`.** A scan of the repo root always exits 3, so CI and
  pre-push scan `apps/api`, and `filet.yml` disables the gate's own douane step. A new Go
  advisory without a severity label also exits 3; the fix is bumping the affected module.
- **Root `biome ci .` processes zero files from a worktree under `~/.cache`** and exits 1. Use
  `pnpm --filter @opus/web exec biome check .`.
- **mise tasks with `depends` fail on mise 2026.9** (`':task' pattern should be expanded`).
  `mise run deps` works; `mise run build`, `test` and `check` do not. Run `go` directly.
- **Back up `opus_go`, not `opus`.** `pg_dump -d opus` silently dumps the retired database.
- The compose file mounts the `opus-uploads` volume at `/app/apps/api/data/uploads`, but the
  entrypoint sets `UPLOAD_DIR=/app/data/uploads`, so uploaded logos live in the container
  layer and go away on redeploy.
