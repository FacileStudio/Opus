# Opus

Self-hosted project management for small teams: projects, kanban boards, tasks, comments and
time tracking behind one login.

A Go API on the suite's tronc chassis serves the data and the built React client from one
binary. Sign-in goes through Registre with porte. Projects and tasks sync with Sablier over the
Antenne bus.

Live at [opus.facile.studio](https://opus.facile.studio).

## What it does

- Workspaces with members and invitations, synced with Registre's shared spaces, plus a
  personal workspace for everyone
- Projects with kanban columns, task priorities, due dates, labels and task relations
- Rich task descriptions, threaded comments, activity history and full-text search
- Per-task time entries and per-project time reporting
- Sign-in through Facile (Registre) only: no local passwords or sign-up. CLI and API tokens
  keep working
- Live collaboration over WebSockets, optionally fanned out across replicas through Redis
- GitHub pull requests linked to tasks, moving them through the board as they open and merge
- GitHub repositories linked to a project, with their open and closed issues in an Issues view
- Two-way project and task sync with Sablier over the Antenne bus
- A remote MCP server at `/mcp`, so Claude, ChatGPT, Cursor or any MCP client can work on your
  projects and tasks, signed in through Registre

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.26, chi, GORM over PostgreSQL 16, goose SQL migrations, tronc, porte, antenne-client |
| Client | React 19, Vite 7, TanStack Router and Query, Tailwind CSS 4, base-ui and Radix, Tiptap 3 |
| Tooling | pnpm 10.32 workspaces and Turborepo for the client, Biome, Vitest, filet and douane for Go |
| Storage | PostgreSQL for data, local disk or S3-compatible storage for task images, local disk for workspace logos |
| Deploy | Docker Compose, one Go container serving the API and SPA on `:80`, Helm chart in `charts/opus` |

## Quick start

```sh
cp .env.sample .env
docker compose up -d
```

Set at least `OPUS_CLIENT_URL`, `CORS_ORIGINS`, `POSTGRES_PASSWORD` and the `OIDC_*` variables
in `.env` before starting: without a Registre issuer nobody can sign in. The compose file
publishes no host ports. The `opus` container listens on `80` internally and is meant to sit
behind Traefik; add a `ports:` mapping to reach it directly.

### Local development

```sh
pnpm install --frozen-lockfile
cd apps/api && go run .                  # the API on :1337, reads apps/api/.env
pnpm --filter @opus/web dev              # the client on :5173
```

Migrations run when the API starts. See [docs/development.md](docs/development.md).

## Command line

[`opus`](https://github.com/FacileStudio/opus-cli) drives Opus from a terminal. It
installs through [facile](https://github.com/FacileStudio/facile), the suite installer,
which fetches a prebuilt binary and puts `~/.local/bin` on your PATH:

```sh
curl -fsSL https://get.facile.studio | bash
facile install opus
```

`opus login` prints a browser approval link and saves a session token. Personal API tokens are
also listed and minted with `opus token`, or in Settings, Account, Developer.

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
| `OPUS_CLIENT_URL` | Public URL of the instance, written into the built client at container start |
| `CORS_ORIGINS` | Allowed browser origins; the first one is where a login lands |
| `DATABASE_URL` | Postgres connection string; derived from the `POSTGRES_*` variables if unset |
| `POSTGRES_PASSWORD` | Password for the Postgres container, also used to derive `DATABASE_URL` |
| `OIDC_ISSUER` | Registre. Opus has no password or sign-up route, so without it nobody can sign in |
| `OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` / `OIDC_REDIRECT_URL` | Opus's client in Registre. The redirect is `<public URL>/api/auth/oidc/callback` |
| `ANTENNE_URL` / `ANTENNE_SECRET` | The Antenne bus for the Sablier sync. Empty turns it off |
| `MCP_RESOURCE` | The MCP server's public URL, such as `https://opus.facile.studio/mcp`. Needs `OIDC_ISSUER`; unset leaves `/mcp` off |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go API: modules/ per domain, internal/ plumbing, migrations/ goose SQL
  web/       React 19 client: TanStack Router pages, fetchers, query and mutation hooks
packages/
  libs/      TypeScript helpers for the client
charts/      Helm chart for Kubernetes deployments
deploy/      Container entrypoint and Postgres init script
i18n/        Translation files and the scripts that validate them
tests/       TypeScript tests of the retired Hono API; nothing runs them
docs/        Architecture, configuration, development, deployment, API, erasure, Sablier sync
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, data model, sign-in, the Antenne bus |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, checks, database migrations |
| [Deployment](docs/deployment.md) | Releases, Docker Compose, the single container, Helm, Traefik |
| [API](docs/api.md) | HTTP routes by module |
| [Erasure](docs/erasure.md) | What a Registre deletion removes and what it anonymises |
| [Sablier sync](docs/sablier-sync.md) | The project, task and time-entry contract with Sablier |

## License

Opus is licensed under [AGPL-3.0-only](LICENSE). It is derived from
[Kaneo](https://github.com/usekaneo/kaneo), whose MIT notice is kept in
[LICENSE-MIT](LICENSE-MIT); see [NOTICE](NOTICE). A commercial licence is available
from Facile Studio for use that cannot meet the AGPL's terms.

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
