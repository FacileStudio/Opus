# Opus

Self-hosted project management for small teams — projects, kanban boards, tasks, comments and
time tracking behind one login.

Opus is the stack outlier of the suite: `pnpm` instead of `bun`, Hono with OpenAPI instead of
tRPC, Drizzle instead of Prisma, and React instead of SvelteKit. It is also the only repo that
consumes another suite repo directly, pulling `@facile/antenne` from
`github:FacileStudio/antenne-client#ts` to sync projects and tasks over the Antenne event bus.

Live at [opus.facile.studio](https://opus.facile.studio).

## What it does

- Workspaces with members, teams and invitations, backed by `better-auth` organizations
- Projects with kanban columns, task priorities, due dates, labels and task relations
- Rich task descriptions, threaded comments, activity history and full-text search
- Per-task time entries and per-project time reporting
- Sign-in through Facile (Registre) only: no local passwords or sign-up. CLI and API tokens
  keep working
- Live collaboration over WebSockets, optionally fanned out across replicas through Redis
- Outbound integrations for GitHub, Gitea, Discord, Slack, Telegram and generic webhooks
- Two-way project and task sync with other suite apps through the Nook pool
- An MCP server and a device authorization flow so CLIs and agents can act on a workspace

## Stack

| Layer | Tech |
|---|---|
| API | Hono 4, `hono-openapi` 1, Drizzle ORM 0.45, PostgreSQL 16, `better-auth` 1.6, Valibot 1 |
| Client | React 19, Vite 7, TanStack Router and Query, Tailwind CSS 4, Radix UI, Tiptap 3 |
| Runtime | Node 20.19+ with `tsx` in development, Bun 1 in the container; pnpm 10.32 workspaces and Turborepo |
| Storage | PostgreSQL for data, S3-compatible object storage for task images, local disk for workspace logos |
| Deploy | Docker Compose, one Go container serving the API and SPA on `:80`, Helm chart in `charts/opus` |

## Quick start

```sh
cp .env.sample .env
docker compose up -d
```

Set at least `OPUS_CLIENT_URL`, `POSTGRES_PASSWORD` and `AUTH_SECRET` in `.env` before starting.
The compose file publishes no host ports — the `opus` container listens on `80` internally and is
meant to sit behind Traefik. Add a `ports:` mapping to reach it directly.

### Local development

```sh
pnpm install
pnpm dev
```

The API listens on `http://localhost:1337` and the web client on `http://localhost:5173`.
Migrations run automatically when the API starts.

## Command line

[`opus`](https://github.com/FacileStudio/opus-cli) drives Opus from a terminal. It
installs through [facile](https://github.com/FacileStudio/facile), the suite installer,
which fetches a prebuilt binary and puts `~/.local/bin` on your PATH:

```sh
curl -fsSL https://get.facile.studio | bash
facile install opus
```

Opus mints a static API key; `facile login opus` opens the page that
generates it (Settings → Account → Developer).

## Configuration

| Variable | What it does |
|---|---|
| `OPUS_CLIENT_URL` | Public URL of the web client; drives CORS, cookies and email links |
| `OPUS_API_URL` | Public URL of the API; derived from `OPUS_CLIENT_URL` inside the container |
| `DATABASE_URL` | Postgres connection string; derived from the `POSTGRES_*` variables if unset |
| `POSTGRES_PASSWORD` | Password for the Postgres container, also used to derive `DATABASE_URL` |
| `AUTH_SECRET` | `better-auth` signing key, 32 characters or more; a random one is generated if unset |
| `OIDC_ISSUER` | Registre. With an issuer set, Opus mounts no password or sign-up route |
| `APP_VERSION` | Build arg, served at `GET /version` |
| `NOOK_POOL_URL` | Default Nook pool URL used when a workspace has no explicit one |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Hono backend — one directory per feature, drizzle/ holds SQL migrations
  web/       React 19 client — TanStack Router pages, fetchers, query and mutation hooks
  site/      Next.js marketing site
  docs/      Standalone product documentation site
packages/
  email/     Transactional email templates and the SMTP transport
  libs/      Types and helpers shared between api and web
  mcp/       MCP server exposing Opus to agents
charts/      Helm chart for Kubernetes deployments
i18n/        Translation files and the scripts that validate them
tests/       API unit tests and API integration tests
docs/        Architecture, configuration, development, deployment, API
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, data model, auth, the Nook pool integration |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, linting, database migrations |
| [Deployment](docs/deployment.md) | Docker Compose, the single container, Helm, Traefik routing |
| [API](docs/api.md) | HTTP endpoints and payloads |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
