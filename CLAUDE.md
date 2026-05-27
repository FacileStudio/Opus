# Opus

Self-hosted project management platform. Simplicity and performance first -- features solve real problems, nothing more. Do not add features, refactoring, or improvements beyond what was asked.

## Tech Stack

| Layer | Tech |
|-------|------|
| Monorepo | pnpm 10.32 + TurboRepo |
| Backend | Hono, Node/tsx, PostgreSQL, Drizzle ORM, Better Auth, Valibot, OpenAPI |
| Frontend | React 19, Vite, TanStack Router (file-based) + Query, Tailwind CSS v4, Zustand, Radix UI, Tiptap editor |
| Docs | Fumadocs (Next.js) at `apps/docs/` |
| Marketing site | Next.js at `apps/site/` |
| MCP server | `packages/mcp/` |
| Email | `packages/email/` |
| Shared libs | `packages/libs/` |
| Linting | Biome (spaces for JS/TS, tabs elsewhere, double quotes, semicolons required) |
| Commits | Conventional Commits enforced by commitlint + Husky |
| Deploy | Docker (single-image), Helm chart (`charts/`), optional Redis for HA WebSocket pub/sub |
| i18n | JSON translation files in `i18n/`, scripts in `scripts/i18n/` |

## Commands

```bash
pnpm install                          # install deps
pnpm dev                              # start all dev servers (API :1337, web :5173)
pnpm build                            # build everything via Turbo
pnpm lint                             # biome check --write across the monorepo
pnpm test                             # unit tests (Vitest) across packages
pnpm test:integration                 # API integration tests (needs PostgreSQL)

# API-specific
pnpm --filter @opus/api dev           # API dev server only
pnpm --filter @opus/api db:generate   # generate Drizzle migration after schema change
pnpm --filter @opus/api db:migrate    # run migrations (also auto-runs on startup)
pnpm --filter @opus/api db:studio     # Drizzle Studio GUI

# Web-specific
pnpm --filter @opus/web dev           # web dev server only
pnpm --filter @opus/web build         # production build
pnpm --filter @opus/web typecheck     # tsc --noEmit

# i18n
pnpm i18n:check                       # validate translation files
pnpm i18n:check:fix                   # auto-fix translation issues
```

## Project Structure

```
opus/
  apps/
    api/              # Hono backend (@opus/api)
      src/            #   feature dirs: {feature}/controllers/, index.ts
      drizzle/        #   SQL migrations
    web/              # React frontend (@opus/web)
      src/
        routes/       #   file-based TanStack Router pages
        components/   #   UI components
        hooks/        #   queries/, mutations/
        fetchers/     #   API client functions by feature
    docs/             # Fumadocs documentation site
    site/             # Marketing/landing site (Next.js)
  packages/
    email/            # Email utilities
    libs/             # Shared libraries (@opus/libs)
    mcp/              # MCP server
    typescript-config/ # Shared tsconfig
  charts/             # Kubernetes Helm charts
  i18n/               # Translation JSON files
  tests/
    api/              # API unit tests
    api-integration/  # API integration tests
  deploy/             # Entrypoint script for Docker
```

## Architecture Patterns

### Backend (API)

- Routes organized by feature: `apps/api/src/{feature}/index.ts`
- Business logic in controllers: `apps/api/src/{feature}/controllers/`
- All routes use `describeRoute` (OpenAPI) and `validator` with Valibot schemas
- Database schema: `apps/api/src/database/schema.ts`, relations in `relations.ts`
- All tables use CUID2 primary keys, `createdAt`/`updatedAt` timestamps, explicit cascade on FKs
- Migrations auto-run on API startup
- Auth context: `c.get("userId")`, `c.get("user")`, `c.get("session")`
- Events: `publishEvent()` from `apps/api/src/events/` for activity tracking
- Errors: throw `HTTPException`

### Frontend (Web)

- File-based routing in `apps/web/src/routes/`
- Fetchers in `apps/web/src/fetchers/{feature}/`
- Query hooks in `apps/web/src/hooks/queries/`, mutation hooks in `hooks/mutations/`
- TanStack Query for server state, Zustand for client state
- Auth client: `@/lib/auth-client`
- User feedback: sonner toasts
- Rich text: Tiptap

## Code Style

- Biome handles formatting and linting; run `pnpm lint` to auto-fix
- CSS and `package.json` files are excluded from Biome
- Prefer `type` over `interface`
- Prefer type inference when obvious
- File naming: PascalCase for components, kebab-case for utilities/hooks (prefix hooks with `use`)
- Import order: externals, then `@/` aliases, then relative (Biome auto-sorts)
- No inline comments in code

## Pre-commit Hooks

The Husky pre-commit hook runs:
1. `biome ci .` -- lint/format check
2. `pnpm run build` -- full monorepo build

Commits are slow because of the build step. Make sure code compiles before committing.
Commit messages are validated by commitlint (Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`, etc.).

## Environment

Single `.env` in project root (shared by all apps). See `.env.sample` for all variables.

Required:
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` (or `DATABASE_URL` directly)
- `AUTH_SECRET` -- min 32 chars, generate with `openssl rand -hex 32`

Important optional:
- `OPUS_CLIENT_URL` (default http://localhost:5173)
- `OPUS_API_URL` (default derived from client URL)
- `REDIS_URL` -- enables multi-instance WebSocket pub/sub (falls back to in-memory)
- SSO: GitHub, Google, Discord, custom OIDC
- SMTP for transactional email

## Testing

- Unit tests: `pnpm test` (Vitest, runs across all packages via Turbo)
- API integration tests: `pnpm test:integration` (requires running PostgreSQL)
- Vitest configs: `apps/api/vitest.config.ts`, `apps/api/vitest.integration.config.ts`, `apps/web/vitest.config.ts`
- Integration test setup: `tests/api-integration/setup.ts`

## Gotchas

- Package manager is **pnpm** -- not npm, not yarn, not bun. Node >= 20.19.0.
- Migrations auto-run on API startup; no need to manually run them in dev after generating.
- The pre-commit build step means every commit triggers a full build. Budget time accordingly.
- `CORS_ORIGINS` empty = allow all in dev. Set explicitly in production.
- Biome excludes CSS files; Tailwind handles CSS linting/formatting.
- `routeTree.gen.ts` is auto-generated by TanStack Router plugin; do not edit manually.
