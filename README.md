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
- A remote MCP server at `/mcp`, so Claude, ChatGPT, Cursor or any MCP client can work on your
  projects and tasks, signed in through Registre

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

## MCP

Opus serves a remote [MCP](https://modelcontextprotocol.io) server at
`https://opus.facile.studio/mcp` (Streamable HTTP, stateless, spec 2026-07-28). Add that URL to
any MCP client as a remote server. There is nothing to install and no API key to copy.

```sh
claude mcp add --transport http opus https://opus.facile.studio/mcp
```

**Sign-in.** The client finds Registre through the server's
`/.well-known/oauth-protected-resource/mcp` document and signs you in with OAuth (PKCE). Registre
shows a consent screen and issues a token bound to this server, with the scopes `opus.read` and
`opus.write`. Revoke a client under Connected apps in your Facile account. Your first MCP call
creates your Opus account if you never opened the web app. MCP is free on every plan.

**What a tool can see.** Exactly what you see in the web app: the workspaces you belong to, their
projects and tasks. Every write goes through the same checks, live updates and suite sync as the
web app.

| Tool | What it does |
|---|---|
| `opus_list_workspaces` | Your workspaces, their members (for assignees) and labels |
| `opus_list_projects` | Projects with their key and the statuses a task can take |
| `opus_search_tasks` | Find tasks by words, status, priority, assignee, label or due date |
| `opus_read_task` | One task in full, by id or ref such as `web-12`, with its comments |
| `opus_create_task` | Add a task, with assignee, due date and labels |
| `opus_update_task` | Change status, assignee, dates, labels or project; only the fields set change |
| `opus_comment_task` | Comment on a task under your name |
| `opus_create_project` | Create a project with the default board |
| `opus_update_project` | Rename, describe, archive or restore a project |
| `opus_delete_task` | Delete a task, after you confirm |
| `opus_delete_project` | Delete a project and its tasks, after you confirm |

**Confirmations.** Deleting never happens on the agent's word alone. A client that can ask you
(MRTR elicitation) shows a confirmation; any other client gets a link to
`https://opus.facile.studio/pending`, where you confirm or dismiss the request. A task or project
edited after the request cannot be confirmed.

## Configuration

| Variable | What it does |
|---|---|
| `OPUS_CLIENT_URL` | Public URL of the web client; drives CORS, cookies and email links |
| `OPUS_API_URL` | Public URL of the API; derived from `OPUS_CLIENT_URL` inside the container |
| `DATABASE_URL` | Postgres connection string; derived from the `POSTGRES_*` variables if unset |
| `POSTGRES_PASSWORD` | Password for the Postgres container, also used to derive `DATABASE_URL` |
| `AUTH_SECRET` | `better-auth` signing key, 32 characters or more; a random one is generated if unset |
| `OIDC_ISSUER` | Registre. With an issuer set, Opus mounts no password or sign-up route |
| `NOOK_POOL_URL` | Default Nook pool URL used when a workspace has no explicit one |
| `MCP_RESOURCE` | The MCP server's public URL, such as `https://opus.facile.studio/mcp`. Needs `OIDC_ISSUER`; unset leaves `/mcp` off |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Hono backend — one directory per feature, drizzle/ holds SQL migrations
  web/       React 19 client — TanStack Router pages, fetchers, query and mutation hooks
packages/
  libs/      Types and helpers shared between api and web
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
