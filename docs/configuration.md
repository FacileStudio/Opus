# Opus: configuration

Every environment variable Opus reads, grouped by concern.

The Go API (`apps/api/internal/env`) loads `.env` from its working directory through godotenv,
then reads the process environment. A variable that is already set in the environment wins over
`.env`. In the container, `deploy/opus-entrypoint.sh` runs first and derives a few variables
before it starts the binary.

The boot stops with an error in these cases:

- `DATABASE_URL` is empty.
- `PORT` is not a TCP port, or `LOG_LEVEL` is not `debug`, `info`, `warn` or `error`.
- `OIDC_ISSUER` is set without `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET` and `OIDC_REDIRECT_URL`.
- `ANTENNE_URL` is set without `ANTENNE_SECRET`.
- `PORTE_TOKEN_KEY` is set but is not a valid key.

Everything else is optional and turns one feature off when unset.

## Core

| Variable | Default | What it does |
|---|---|---|
| `DATABASE_URL` | none | Postgres connection string. Required. The entrypoint builds it from `POSTGRES_*` when unset |
| `PORT` | `8080`, `80` in the image | Listen port. Local development uses `1337`, where the client's dev build points |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `CORS_ORIGINS` | empty | Allowed browser origins, comma-separated. tronc reads the first one set of `CORS_ALLOWED_ORIGINS`, `ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`, `TRUSTED_ORIGINS` and `CLIENT_ORIGIN`. Production sets `CORS_ORIGINS` |
| `TRUSTED_PROXIES` | private ranges | Proxies whose forwarded headers the router trusts: CIDR blocks, addresses, `private`, `cloudflare` or `none` |
| `CLIENT_DIR` | `./client`, `/client` in the image | Where the built SPA lives. With no client there, the binary serves the API only |
| `UPLOAD_DIR` | `./data/uploads`, `/app/data/uploads` from the entrypoint | Workspace logos and local assets. The compose file mounts the `opus-uploads` volume there |
| `JOURNAL_URL` / `JOURNAL_TOKEN` | empty | Ship logs to Journal when both are set |

## Container entrypoint

`deploy/opus-entrypoint.sh` reads these before it starts the binary. The Go API does not read
them.

| Variable | Default | What it does |
|---|---|---|
| `POSTGRES_DB` | `opus_go` | Database name used to build `DATABASE_URL` |
| `POSTGRES_USER` | `opus` | Same |
| `POSTGRES_PASSWORD` | none | Same. The entrypoint exits 1 when both `DATABASE_URL` and this are unset |
| `POSTGRES_HOST` | `postgres` | Same |
| `POSTGRES_PORT` | `5432` | Same |
| `GO_DATABASE_URL` | empty | Replaces `DATABASE_URL` when set |
| `OPUS_CLIENT_URL` | empty | Public URL of the instance. `apps/web/env.sh` writes it into the built JavaScript |
| `OPUS_API_URL` | `<OPUS_CLIENT_URL>/api` | Written into the built JavaScript in place of the `VITE_API_URL` placeholder |

The entrypoint URL-encodes the user and password when the image has Node or Bun. The runtime
image has neither, so it inserts them as they are: keep them free of `@`, `:`, `/` and `%`, or set
`DATABASE_URL` yourself.

The compose file shares one `.env` between the `postgres` and `opus` services, so `POSTGRES_DB`
there names both the database Postgres creates on a fresh volume and the one the app connects
to. `deploy/db/init/01-opus-go.sql` also creates `opus_go` on a fresh volume. Production sets
`POSTGRES_DB=opus_go`; the `opus` database in the same container is the retired TypeScript
app's.

At boot the API creates the database named in `DATABASE_URL` when it is missing and the role may
do so, then applies the goose migrations.

## Sign-in

Opus signs people in through Registre only, with `porte/oidc`. It has no password, sign-up or
magic-link route, so without `OIDC_ISSUER` nobody can sign in.

| Variable | Default | What it does |
|---|---|---|
| `OIDC_ISSUER` | empty | Registre, `https://sso.facile.studio` in production. Empty turns sign-in off |
| `OIDC_CLIENT_ID` | empty | Required with an issuer. `opus` in production |
| `OIDC_CLIENT_SECRET` | empty | Required with an issuer |
| `OIDC_REDIRECT_URL` | empty | Required with an issuer: `<public URL>/api/auth/oidc/callback`. Without the `/api` prefix the callback lands on the SPA and the login never completes |
| `OIDC_SUCCESS_URL` | first CORS origin | Where a completed login lands |
| `OIDC_CLI_AUDIENCE` | empty | Mounts porte's device exchange (`POST /api/auth/oidc/exchange`) for the suite CLIs. Empty answers 404 |
| `OIDC_MACHINE_AUDIENCE` | empty | Opus's own client id. Turns on offline verification of Registre machine JWTs for other suite apps, such as Bulle |
| `OIDC_SCOPE_PREFIX` | empty | Set to `opus`. A machine JWT then needs `opus.read` for GET, HEAD and OPTIONS and `opus.write` for everything else |
| `PORTE_TOKEN_KEY` | empty | Seals the OIDC tokens porte stores. One or more comma-separated keys, newest first, each 64 hex characters (`openssl rand -hex 32`) or base64 of 32 bytes. When set, the boot reseals existing rows in the background |

Sessions and personal API tokens are porte sessions, so they need no secret of their own.

## MCP server

| Variable | Default | What it does |
|---|---|---|
| `MCP_RESOURCE` | empty | Public URL of the MCP endpoint, such as `https://opus.facile.studio/mcp`. With `OIDC_ISSUER` set it turns `/mcp` on. Registre must declare the same URL as Opus's MCP resource |

The key that signs confirmation questions is derived from `OIDC_CLIENT_SECRET`, so it needs no
variable.

## Ideas

| Variable | Default | What it does |
|---|---|---|
| `BULLE_URL` | empty | Bulle's base URL. With `OIDC_ISSUER` set it mounts the `/api/ideas` routes and the Ideas page |

## Antenne event bus

Opus publishes project and task events to Antenne and applies the ones Sablier publishes. With
no URL the bus is off and local writes are unaffected. See [Sablier sync](sablier-sync.md).

| Variable | Default | What it does |
|---|---|---|
| `ANTENNE_URL` | empty | Antenne instance URL |
| `ANTENNE_INSTANCE` | empty | Older name for `ANTENNE_URL`. `ANTENNE_URL` wins when both are set |
| `ANTENNE_SECRET` | empty | Shared secret of the instance. Required with a URL |
| `ANTENNE_INSTANCE_ID` | empty | Distinguishes several Opus deployments on one bus |
| `ANTENNE_SYNC_WORKSPACE_ID` | empty | Fallback workspace for projects from old Sablier versions whose events carry no `space` |
| `ANTENNE_APP` | `Opus` | Overrides the app name Opus registers under. Production leaves it unset |

## GitHub

| Variable | Default | What it does |
|---|---|---|
| `GITHUB_WEBHOOK_SECRET` | empty | Verifies `X-Hub-Signature-256` on `POST /api/integrations/github/webhook` |

The same events can also arrive over the Antenne bus (`git.pull_request`, `git.push`), which
needs no variable here.

## Realtime

| Variable | Default | What it does |
|---|---|---|
| `REDIS_URL` | empty | Redis pub/sub for WebSocket broadcasts across replicas. Empty keeps them in memory, which is only correct for one container |

## Object storage

Task images go to S3-compatible storage through presigned URLs when the first four variables
are all set. Without them the API takes the upload itself (`PUT /api/asset/upload/*`) and writes
it under `UPLOAD_DIR/assets`, with a 10 MiB ceiling. Workspace logos always stay on the local
disk, under `UPLOAD_DIR/workspaces`.

| Variable | Default | What it does |
|---|---|---|
| `S3_ENDPOINT` | empty | S3-compatible endpoint |
| `S3_BUCKET` | empty | Bucket name |
| `S3_ACCESS_KEY_ID` | empty | Access key |
| `S3_SECRET_ACCESS_KEY` | empty | Secret key |
| `S3_REGION` | `us-east-1` | Region |
| `S3_PUBLIC_BASE_URL` | empty | Base URL for publicly readable objects |
| `S3_FORCE_PATH_STYLE` | `true` | Path-style addressing. Any value other than `true` turns it off |
| `S3_MAX_IMAGE_UPLOAD_BYTES` | `10485760` | Upload size ceiling with S3, 10 MiB |
| `S3_PRESIGN_TTL_SECONDS` | `300` | Presigned URL lifetime |

## Web client build

Vite reads these at build time. The production build (`apps/web/.env.production`) sets
`VITE_API_URL` to the placeholder string `OPUS_API_URL`, which `apps/web/env.sh` replaces with the
real value when the container starts.

| Variable | Default | What it does |
|---|---|---|
| `VITE_API_URL` | `http://localhost:1337` | API base URL for fetchers, auth and the WebSocket. Build every URL with `src/fetchers/get-api-url.ts`, which accepts a base with or without `/api` |
| `VITE_APP_URL` | the page's own origin | Base URL of shareable links |

`env.sh` also substitutes any other `OPUS_`-prefixed variable it finds in the container
environment into the built JavaScript and CSS.
