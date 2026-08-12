# Opus backend → Go port plan

Cold-start handoff. A fresh session with no prior conversation can pick up any phase
from this document. Approved by saravenpi 2026-08-12; work is on branch `feat/go-backend`.

## Goal
Replace the TypeScript/Hono Opus backend with a Go backend at `apps/api`, on the Facile
stack (**tronc + porte + journal + casier + filet**), preserving the React frontend and its
HTTP contract through a strangler-fig transition.

## Why
The suite consolidated the plumbing Opus re-implements (tronc = chassis, porte = auth,
journal = logging) and the ROADMAP declares Opus "excluded on purpose… being rewritten in
Go". `apps/api-legacy/src/` is ~31k LOC TS across 25 feature modules; better-auth is the
suite's odd-one-out identity engine.

## Approach
Go API at `apps/api` on tronc + porte + journal + goose (tronc/migrate), reusing the Sablier
module pattern (`modules/<feature>/{router,controller,service,types,documentation}.go`).
The old TS API sits at `apps/api-legacy` and stays buildable during the transition; the Go
container serves the SPA + `/api/*` and proxies unported routes to the legacy API. Features
flip Go-native one module at a time. The frontend changes once — when auth swaps from
better-auth to porte.

## Current state (Phase 0, mostly done — commit `…`)
- `git mv apps/api apps/api-legacy` (history preserved); `apps/api-legacy/package.json`
  renamed to `@opus/api-legacy`; root `package.json` `test:integration` repointed.
- New Go module `github.com/FacileStudio/Opus/apps/api`: `main.go` (tronc: httpx chain,
  health, apiref `/docs`, spa, journal-signal shutdown), `internal/env` (tronc Core +
  godotenv), `internal/documentation` (apiref registry types), `mise.toml`, `filet.yml`
  (root, Go-scoped). Builds; `filet check .` exit 0; boots and serves `/health`, `/ready`,
  `/docs`.

## Stack (from go.mod)
tronc v0.13.0, Journal sdk/journal (latest), godotenv. Phase 1 adds porte, chi/gorilla,
gorm (postgres), trivia enveloppe + antenne-client for events.

## Phases (parallel tracks K1 API / K2 frontend-auth / K3 ops)

- **Phase 0 — Foundation** (most done): rename, Go skeleton, filet, mise, Docker dual-run,
  casier pull. Remaining: dual-run Docker (Go + legacy, single Traefik router → Go, proxy
  unported `/api/*` to legacy), goose baseline vs a **fresh** DB (Drizzle stops auto-running).
- **Phase 1 — Auth seam + workspace**: porte (oidc/local/session/pg) replaces better-auth;
  frontend auth re-plumb to porte (`@/lib/auth-client`, `auth-provider`, login, `X-Facile-CSRF`);
  auth data migration (better-auth tables → `porte_identities`, users preserved by email,
  argon2id hashes compatible); port `workspace` + `workspace-user`.
- **Phase 2 — Core domain**: project, column, task (incl. bulk/import/export/number), label,
  task-relation, comment, activity. Parity vs captured OpenAPI per flip.
- **Phase 3 — Collaboration**: invitation, share-link, workspace-webhook, notification +
  preferences, time-entry, search, workflow-rule.
- **Phase 4 — Events + realtime + scheduler**: `events` (enveloppe envelope, keyed on
  `actor_email`) over **antenne-client/go** (NOT "nook pool" — that name is dead; it was
  renamed to Antenne 2026-08-07; `apps/api-legacy/src/nook-pool` and the `nookPool*` tables
  are the legacy names the port scrubs). WS broadcast (gorilla + Redis pub/sub), scheduler
  (due-date reminders).
- **Phase 5 — Integrations**: github, gitea, slack, discord, telegram, generic-webhook,
  plugins.
- **Phase 6 — Decommission**: storage (S3/local presigned) in Go; delete `apps/api-legacy`,
  Drizzle, node backend; single container; Dokploy deploy.

### Casier (design decision — flagged, not resolved)
"Use casier" is read as: pull env secrets from Casier at boot via a machine token (endpoint
`GET /api/projects/{project}/environments/{env}/secrets/export`). No suite Go app consumes
casier yet and it has no Go SDK → hand-rolled client in `internal/casier`. Open question:
must it fail-hard if Casier is unreachable? Deferred until confirmed.

## Exit criteria (per step)
`go build ./...` clean · `filet check .` exit 0 (never adjust thresholds) · `mise run test`
green · parity vs captured OpenAPI where a feature flips · porte login in the browser +
CSRF enforced · events reach Antenne · single container serves SPA + `/api`.

## Risks / unknown unknowns
- **Live-DB auth migration** is the riskiest step (better-auth → porte tables); consider a
  fresh DB at cutover if live data preservation doesn't matter.
- **Casier** ambiguity (boot dependency; fail-hard semantics) unresolved.
- Contract drift between parity captures and the legacy API.
- The `nook.pool`/`nookPool*` legacy names in `apps/api-legacy` are the same event bus as
  Antenne — do not treat them as a separate thing.

## Skip (YAGNI)
- `packages/mcp`, `packages/email` stay TS (separate components, not the API backend).
- `apps/docs`, `apps/site` untouched.
- No better-auth *session* rows migrated at cutover — sessions are short-lived.