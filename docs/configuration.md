# Opus — Configuration

Every environment variable the code actually reads, grouped by concern. Opus loads a single `.env`
at the repo root through `dotenv-mono`, shared by all workspaces.

Opus does not use `tronc`, so there is no `troncenv.Core` block and no fail-fast on a missing
variable. The API boots with almost everything unset and degrades feature by feature — the one
hard exit is an `AUTH_SECRET` shorter than 32 characters.

## Core

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OPUS_CLIENT_URL` | yes | `http://localhost:5173` | Public URL of the web client. Drives CORS fallback, `better-auth` trusted origins, email links and the device verification page |
| `OPUS_API_URL` | no | `http://localhost:1337` | Public URL of the API. The container derives it as `<OPUS_CLIENT_URL>/api` when unset |
| `CORS_ORIGINS` | no | — | Comma-separated allowed origins. Falls back to `OPUS_CLIENT_URL`; when neither is set the request origin is echoed back |
| `COOKIE_DOMAIN` | no | — | Explicit cookie domain, for cross-subdomain deployments |
| `NODE_ENV` | no | `production` in the image | Standard Node environment flag |

## Database

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DATABASE_URL` | yes | — | Postgres connection string. `drizzle-kit` falls back to `postgresql://opus_user:opus_password@localhost:5432/opus` for CLI commands only |
| `POSTGRES_DB` | no | `opus` | Used by the container entrypoint to derive `DATABASE_URL` |
| `POSTGRES_USER` | no | `opus` | Same |
| `POSTGRES_PASSWORD` | yes | — | Same. The entrypoint exits 1 when both `DATABASE_URL` and this are unset |
| `POSTGRES_HOST` | no | `postgres` | Same |
| `POSTGRES_PORT` | no | `5432` | Same |

The entrypoint URL-encodes the user and password before assembling `DATABASE_URL`.

## Authentication

| Variable | Required | Default | What it does |
|---|---|---|---|
| `AUTH_SECRET` | yes | random per boot | `better-auth` signing key. A value under 32 characters exits 1. When unset the container generates one, so sessions do not survive a restart |
| `SSO_ONLY` | no | `false` | `true` rejects email sign-in and sign-up, magic links, email OTP and anonymous sign-in |
| `DISABLE_REGISTRATION` | no | `false` | `true` allows sign-up only with a valid invitation |
| `DISABLE_PASSWORD_REGISTRATION` | no | `false` | `true` rejects `/sign-up/email` but leaves other methods open |
| `DISABLE_GUEST_ACCESS` | no | `false` | `true` removes the `anonymous` plugin |
| `DEMO_MODE` | no | `false` | Reported by `GET /api/config` as `isDemoMode` |
| `DEVICE_AUTH_CLIENT_IDS` | no | `opus-cli,opus-mcp` | Comma-separated client ids accepted by the device authorization flow |

## OAuth and SSO providers

| Variable | Required | Default | What it does |
|---|---|---|---|
| `GITHUB_OAUTH_CLIENT_ID` | no | — | Sign in with GitHub. Preferred over the legacy pair |
| `GITHUB_OAUTH_CLIENT_SECRET` | no | — | Same |
| `GITHUB_CLIENT_ID` | no | — | Legacy fallback, used only when `GITHUB_OAUTH_*` is unset |
| `GITHUB_CLIENT_SECRET` | no | — | Same |
| `GOOGLE_CLIENT_ID` | no | — | Sign in with Google |
| `GOOGLE_CLIENT_SECRET` | no | — | Same |
| `DISCORD_CLIENT_ID` | no | — | Sign in with Discord |
| `DISCORD_CLIENT_SECRET` | no | — | Same |
| `CUSTOM_OAUTH_CLIENT_ID` | no | — | Generic OIDC provider. This is how Opus federates to Authentik |
| `CUSTOM_OAUTH_CLIENT_SECRET` | no | — | Same |
| `CUSTOM_OAUTH_DISCOVERY_URL` | no | — | OIDC discovery document, e.g. the Authentik application's `.well-known` URL |
| `CUSTOM_OAUTH_AUTHORIZATION_URL` | no | — | Explicit authorization endpoint when discovery is not used |
| `CUSTOM_OAUTH_TOKEN_URL` | no | — | Explicit token endpoint |
| `CUSTOM_OAUTH_USER_INFO_URL` | no | — | Explicit userinfo endpoint |
| `CUSTOM_OAUTH_SCOPES` | no | `openid,profile,email` | Comma-separated scopes |
| `CUSTOM_OAUTH_RESPONSE_TYPE` | no | `code` | OAuth response type |
| `CUSTOM_OAUTH_LOGOUT_URL` | no | — | Returned by `GET /api/config`; the client redirects here on sign-out |
| `CUSTOM_AUTH_PKCE` | no | `true` | Set to `false` to disable PKCE. Note the name: `CUSTOM_AUTH_`, not `CUSTOM_OAUTH_` |

Unlike the Go and `jose` apps in the suite, Opus reads none of `OIDC_ISSUER`, `OIDC_CLIENT_ID` or
`OIDC_CLIENT_SECRET`. Federating to `porte.facile.studio` means filling in the `CUSTOM_OAUTH_*`
block.

### Delegated suite access

The Go API verifies short-lived bearer JWTs minted by Registre for other suite services.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `OIDC_MACHINE_AUDIENCE` | no | — | Opus's own client id. Enables offline JWT verification. Requires `OIDC_ISSUER` |
| `OIDC_SCOPE_PREFIX` | no | — | Set to `opus`. A JWT then needs `opus.read` for GET, HEAD and OPTIONS and `opus.write` for everything else. Requires `OIDC_MACHINE_AUDIENCE` |

Opaque API tokens and cookie sessions are untouched by both settings.

## GitHub App integration

Separate from GitHub sign-in. These power repository sync and webhook delivery.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `GITHUB_APP_ID` | no | — | GitHub App id |
| `GITHUB_APP_NAME` | no | — | App slug, used to build installation URLs |
| `GITHUB_PRIVATE_KEY` | no | — | App private key, PEM |
| `GITHUB_WEBHOOK_SECRET` | no | — | Shared secret verifying `POST /api/github-integration/webhook` |

## Antenne event bus

Opus publishes project and task events to Antenne and applies the ones Sablier publishes. All
three variables are optional. With `ANTENNE_URL` or `ANTENNE_SECRET` unset, the bus is off and
local writes are unaffected. See [Sablier sync](sablier-sync.md).

| Variable | Required | Default | What it does |
|---|---|---|---|
| `ANTENNE_URL` | no | — | Antenne instance URL |
| `ANTENNE_SECRET` | no | — | Shared secret of the Antenne instance |
| `ANTENNE_INSTANCE_ID` | no | — | Distinguishes several Opus instances on one bus |
| `ANTENNE_SYNC_WORKSPACE_ID` | no | — | Fallback workspace for projects from old Sablier versions whose events carry no `space`. Current Sablier projects land in their Registre or personal workspace |

## Email

Transactional email is skipped entirely when `SMTP_HOST` or `SMTP_FROM` is unset; invitations are
still created and a warning is logged.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `SMTP_HOST` | no | — | SMTP server hostname |
| `SMTP_PORT` | no | — | SMTP port |
| `SMTP_USER` | no | — | SMTP username |
| `SMTP_PASSWORD` | no | — | SMTP password |
| `SMTP_FROM` | no | — | From address on every outgoing message |
| `SMTP_SECURE` | no | `true` | Implicit TLS. Only the literal string `false` disables it |
| `SMTP_REQUIRE_TLS` | no | `false` | Require STARTTLS |
| `SMTP_IGNORE_TLS` | no | `false` | Skip TLS entirely |

`GET /api/config` reports `hasSmtp` only when `SMTP_HOST`, `SMTP_PORT`, `SMTP_SECURE`, `SMTP_USER`
and `SMTP_PASSWORD` are all set.

## Redis

Optional. Enables WebSocket pub/sub across replicas; without it the broadcast adapter is
in-memory, which is only correct for a single container. Precedence is cluster, then sentinel,
then standalone.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `REDIS_URL` | no | — | Standalone connection string |
| `REDIS_SENTINELS` | no | — | Comma-separated `host:port` sentinel list |
| `REDIS_SENTINEL_MASTER_NAME` | no | `mymaster` | Sentinel master name |
| `REDIS_SENTINEL_PASSWORD` | no | — | Sentinel auth password |
| `REDIS_SENTINEL_TLS` | no | `false` | `true` enables TLS for sentinel mode |
| `REDIS_CLUSTER_NODES` | no | — | Comma-separated `host:port` cluster node list |
| `REDIS_PASSWORD` | no | — | Password for sentinel and cluster modes |

## Object storage

Task image uploads require all four of the first block; calling an upload route without them
throws. Workspace logos never touch S3.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `S3_ENDPOINT` | for uploads | — | S3-compatible endpoint |
| `S3_BUCKET` | for uploads | — | Bucket name |
| `S3_ACCESS_KEY_ID` | for uploads | — | Access key |
| `S3_SECRET_ACCESS_KEY` | for uploads | — | Secret key |
| `S3_REGION` | no | `us-east-1` | Region |
| `S3_PUBLIC_BASE_URL` | no | — | Base URL for publicly readable objects |
| `S3_FORCE_PATH_STYLE` | no | `true` | Path-style addressing, required by most self-hosted S3 |
| `S3_MAX_IMAGE_UPLOAD_BYTES` | no | `10485760` | Upload size ceiling, 10 MiB |
| `S3_PRESIGN_TTL_SECONDS` | no | `300` | Presigned URL lifetime |
| `UPLOAD_DIR` | no | `<cwd>/data/uploads` | Local directory for workspace logos. The compose file mounts `opus-uploads` here |

## Nook pool

| Variable | Required | Default | What it does |
|---|---|---|---|
| `NOOK_POOL_URL` | no | — | Default pool URL, used when a workspace's `nook_pool_config` row leaves `instanceUrl` blank |
| `NOOK_POOL_SECRET` | no | — | Default pool secret, same fallback rule |

Neither variable enables the pool on its own. A workspace also needs an enabled `nook_pool_config`
row, created through `POST /api/nook-pool/:workspaceId`.

## Notifications

| Variable | Required | Default | What it does |
|---|---|---|---|
| `NOTIFICATION_SECRET_ENCRYPTION_KEY` | for encrypted secrets | — | Key used to encrypt provider secrets stored in notification preferences. Storing one without this set throws |

## Web client build

Read at build time by Vite and baked into the bundle, then rewritten at container start by
`apps/web/env.sh`, which substitutes the literal strings `OPUS_API_URL` and `OPUS_CLIENT_URL` in
the emitted JavaScript.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `VITE_API_URL` | no | `http://localhost:1337` | API base URL used by fetchers, the auth client and the WebSocket hook |
| `VITE_CLIENT_URL` | no | — | Base URL for OAuth callback and error redirects |
| `VITE_APP_URL` | no | — | Base URL used when generating shareable links |

Any other `OPUS_`-prefixed variable present in the container environment is also substituted into
the built assets by `env.sh`.

## MCP server

Read by `packages/mcp` when it runs as a client-side process, not by the API.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OPUS_MCP_CLIENT_ID` | no | `opus-mcp` | Device-flow client id. Must appear in `DEVICE_AUTH_CLIENT_IDS` on the server |
| `XDG_CONFIG_HOME` | no | `~/.config` | Where the MCP token store writes credentials |
| `APPDATA` | no | — | Windows equivalent, used to resolve the Claude Desktop config path |
