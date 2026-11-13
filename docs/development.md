# Opus: development

Local setup, the commands you will use, how migrations work, and the checks that run before a
commit, a push and a merge.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.26.11 | The `go` line in `apps/api/go.mod`. An older local Go downloads it through `GOTOOLCHAIN` |
| Node | 22 (20.19 at least) | `mise.toml` pins 22.22.1; `engines` in the root `package.json` asks for 20.19 or later, which Vite 7 needs |
| pnpm | 10.32.1 | Pinned as `packageManager`. Not npm, not yarn, not bun: `bun install` rewrites the root `package.json` workspaces |
| PostgreSQL | 16 | The API runs on Postgres only, tests included |
| Docker | any recent | For a throwaway Postgres |

## Setup

```sh
pnpm install --frozen-lockfile
docker run -d --name opus-dev -e POSTGRES_PASSWORD=opus -p 127.0.0.1:5433:5432 postgres:16
```

The API reads `apps/api/.env`. A minimal one for local work:

```sh
DATABASE_URL=postgres://postgres:opus@127.0.0.1:5433/opus_go?sslmode=disable
PORT=1337
CORS_ORIGINS=http://localhost:5173
```

The API creates the `opus_go` database at boot when it is missing. Sign-in needs a Registre
client too: set `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and an
`OIDC_REDIRECT_URL` of `http://localhost:1337/api/auth/oidc/callback` that the client declares.
Without them the API runs but nobody can sign in. Every variable is in
[configuration.md](configuration.md).

## Run

Two processes, in two terminals:

```sh
cd apps/api && go run .              # the API on :1337
pnpm --filter @opus/web dev          # the client on :5173
```

`pnpm dev` runs Turbo's `dev` task, which starts the client only: `apps/api` has no
`package.json`. The client's development build calls `http://localhost:1337`
(`apps/web/.env.development`).

`mise run dev` in `apps/api` fails on mise 2026.9 with `':task' pattern should be expanded`, as
does every task there that has `depends`. Run `go` directly.

The API reference is at `http://localhost:1337/docs`, and the raw spec at `/docs/openapi.json`.

## Database migrations

Migrations are ordered goose SQL files in `apps/api/migrations/`, embedded in the binary and
applied at every boot by `internal/database.RunMigrations`, under a Postgres session lock so two
replicas never migrate at once. GORM is a query builder here: `AutoMigrate` never runs.

To change the schema:

1. Add the next numbered file, such as `00010_<what>.sql`, with `-- +goose Up` and
   `-- +goose Down`.
2. Update the matching struct in `apps/api/schemas/` by hand.
3. Run the tests against Postgres: each test package gets a schema built from the real
   migrations.

goose must stay at v3.27.3 or later: older versions look for their ledger in `public` and fail in
the per-schema test databases.

## Tests

### API

```sh
docker run -d --name opus-test -e POSTGRES_PASSWORD=opus -p 127.0.0.1:5434:5432 \
  postgres:16 -c fsync=off -c synchronous_commit=off -c full_page_writes=off
cd apps/api
TEST_DATABASE_URL='postgres://postgres:opus@127.0.0.1:5434/postgres?sslmode=disable' go test -race ./...
```

`TEST_DATABASE_URL` is a full URL with a database name. Tests open it through
`internal/database/dbtest` (`tronc/testdb`): one schema per package, empty tables for every
test. Without the variable, database tests skip and the rest still run. `fsync=off` matters:
each test truncates every table, which costs about 3 seconds per test with fsync on.

### Client

```sh
pnpm --filter @opus/web test          # Vitest
pnpm --filter @opus/web typecheck     # tsc --noEmit
pnpm --filter @opus/web exec biome check .
pnpm i18n:check
```

`tests/` holds TypeScript tests for the retired Hono API. Nothing runs them.

## Checks

### Git hooks

`lefthook.yml` pulls the shared commit-msg policy from `FacileStudio/hooks` (`ref: v1`) and adds:

- `pre-commit`: `pnpm exec biome ci .`, then `pnpm run build`. The full build is deliberate and
  slow. It needs Node 20.19 or later on `PATH`; an older one fails with
  `crypto.hash is not a function`.
- `pre-push`: `filet gate .` (gofmt, the rules in `filet.yml`, vet, staticcheck and the tests),
  then `douane scan -fail critical -scope prod apps/api`. douane scans `apps/api` only because it
  cannot read `pnpm-lock.yaml`, and `filet.yml` turns the gate's own douane step off for the
  same reason. Export `TEST_DATABASE_URL` first, or the Go database tests skip.
- `commit-msg`: validates Conventional Commits and prepends the gitmoji. Never write the emoji
  yourself.

`mise install` installs the hooks (`[hooks] postinstall = "lefthook install"` in `mise.toml`).
A clone without them commits with no gitmoji and no message check; run the checks above by hand.

### CI

`.github/workflows/ci.yml` runs on every push to `main` and every pull request:

- `api`: rejects a non-Postgres GORM driver in `go.mod`, then gofmt, `go vet`, `go build` and
  `go test -race` against a Postgres 16 started with `fsync=off`.
- `web`: `pnpm i18n:check`, typecheck, `biome check`, Vitest and a production build.
- `filet` and `douane`, as in the pre-push hook.

`release.yml` runs on a `v*` tag. See [deployment.md](deployment.md).

## Formatting

Go: `gofmt`. The client: Biome (`biome.json`), with spaces for JavaScript and TypeScript, tabs
elsewhere, double quotes, required semicolons and sorted imports. CSS and `package.json` are
excluded.

```sh
gofmt -w apps/api
pnpm lint                             # biome check --write in every workspace
```

## Internationalization

Translations are JSON files in `i18n/`: `de-DE`, `el-GR`, `en-US`, `es-ES`, `fr-FR`, `mk-MK`,
`nl-NL`, `ru-RU` and `uk-UA`, validated against `i18n/schema.json`.

```sh
pnpm i18n:check       # validate
pnpm i18n:check:fix   # fix what can be fixed
pnpm i18n:report      # coverage per locale
pnpm i18n:schema      # regenerate the schema
```

## Conventions

API (`apps/api`):

- One package per domain under `modules/<name>/`: `types.go`, `service.go`, `controller.go`,
  `router.go` and `documentation.go`. Shared plumbing lives under `internal/`.
- Routes register inside `router.Route("/api", ...)` in `router.go`. Only `/health`, `/ready`,
  `/version`, `/docs`, `/mcp` and `/ws` live at the root, and the SPA catch-all is mounted last.
- Document every route in the module's `documentation.go`. `referenceConfig()` collects them for
  `/docs`, and `router_test.go` fails when a module is missing.
- Request IDs use `internal/wireid`, which decodes a JSON number or a decimal string. The client
  sends IDs as strings and tronc decodes strictly, so a plain `int64` field answers 400.
- New in-process callers go through a module controller's `Calls()`, as the MCP server does. The
  services' raw methods skip realtime broadcasts and Antenne events.
- Events go through the outbox: `internal/events` enqueues in the same transaction as the change.
  A new payload type needs a case in `canonical()` in `internal/events/outbox.go`.
- Errors are tronc `errors` values, written with `httpjson.WriteError`.

Client (`apps/web`):

- Build every API URL with `src/fetchers/get-api-url.ts`. A hand-built base produced `/api/api`.
- Fetchers in `src/fetchers/<feature>/`, query hooks in `src/hooks/queries/`, mutation hooks in
  `src/hooks/mutations/`. TanStack Query owns server state, Zustand owns client state.
- `src/routeTree.gen.ts` is generated by the TanStack Router plugin. Do not edit it.

Everywhere: no inline comments in code.
