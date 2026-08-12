# Opus backend → Go port plan

Cold-start handoff. A fresh session with NO prior conversation can pick up any phase from
this document. Branch `feat/go-backend`. Commits with `git commit --no-verify` (husky runs
biome + full pnpm build; see Gotchas).

## Goal
Replace the TypeScript/Hono Opus backend with a Go backend at `apps/api`, on the Facile
stack (**tronc + porte + journal + casier + filet**), preserving the React frontend and its
HTTP contract through a strangler-fig transition.

## Why
The suite consolidated the plumbing Opus re-implements (tronc chassis, porte auth, journal
logging) and the ROADMAP declares Opus "excluded on purpose… being rewritten in Go".
`apps/api-legacy/src/` is ~31k LOC TS across 25 feature modules; better-auth is the suite's
odd-one-out identity engine.

## Approach
Go API at `apps/api` on tronc + porte + journal (+ casier for secrets), reusing the Sablier
module pattern (`modules/<feature>/{router,controller,service,types,documentation}.go`).
The old TS API lives at `apps/api-legacy` and keeps running; the Go binary is the single
router (serves the SPA at `/client` + `/api/*`) and proxies unported routes to the legacy
API at `LEGACY_API_URL` (127.0.0.1:1337). Features flip Go-native one module at a time.
The frontend changes once — when auth swaps from better-auth to porte (Phase 1).

## Decisions (baked in — re-open only if something breaks against prod)
- **Fresh DB for the port dev cycle.** The Go backend and the legacy API share one dev
  Postgres that a `docker compose up` owns. Do NOT write a better-auth→porte data migration
  now. Live prod keeps running on the old TS API until Phase 6.
- **Schema ownership stays with Drizzle until Phase 6.** Both apps share one dev DB; the
  legacy API auto-runs Drizzle migrations on boot and they are the schema authority. The Go
  backend must NOT run goose against the same DB — dual ownership of one schema is the bug
  class to avoid. The goose baseline (`apps/api/migrations/00001_baseline.sql` + own package
  with `//go:embed *.sql`) is a **Phase 6** step, taken from a `pg_dump --schema-only` of the
  live DB at cutover, with the raw `INSERT INTO goose_db_version` baseline dance.
- **Casier = fail-soft.** `CASIER_MACHINE_TOKEN` set → pull env secrets from Casier at boot
  (`GET /api/projects/{project}/environments/{env}/secrets/export` with the machine token);
  on any failure, log and continue with `.env`/env vars. Hard fail only once prod cutover is
  proven. No Go SDK exists — hand-rolled client in `apps/api/internal/casier`.
- **`@opus/api` package name is KEPT** (directory moved to `apps/api-legacy`, name did not):
  `packages/libs` (`workspace:*`, `AppType` import) and the test suite depend on the name.
  The pnpm-lockfile importer was hand-renamed `apps/api` → `apps/api-legacy` (2 lines).

## Current state — Phase 0 DONE (commits `1c3b48f7`, `7cf0d993`, `15a4e3dd`)
- TS API at `apps/api-legacy` (git mv, history preserved); package `@opus/api`; tests'
  path imports retargeted; lockfile importer renamed.
- Go module `github.com/FacileStudio/Opus/apps/api` (go 1.26): tronc chain (httpx, health,
  apiref `/docs`, spa, journal-signal shutdown), `internal/env` (Core + `LEGACY_API_URL`),
  `internal/documentation` (apiref types), `internal/legacy` (strangler proxy under `/api`,
  WebSocket upgrades pass through), `mise.toml`, `filet.yml` (root, Go-scoped, `.go` only;
  biome owns the TS).
- Docker single-container dual-run: Go binary on :80 serves `/client` (SPA) + `/api/*`,
  proxies unported paths to legacy on 127.0.0.1:1337. nginx gone. Legacy runtime stays at
  `/app/apps/api` (uploads/migrations path preserved) with `node_modules` symlinked from the
  workspace package (`apps/api-legacy/node_modules`) — pnpm hoisting bridge. `web/env.sh`
  retargeted to `CLIENT_DIR` for runtime URL injection. `.dockerignore` at root.
- Gates: `go build`/`vet`/`test` green; `filet check .` exit 0; `filet docker .` warns only
  (RUN-layering + pnpm-as-npm false positives — ignore, don't tune config).

## Reference implementation (read before writing Phase 1 code)
- **Sablier** `apps/api/main.go` — the canonical tronc+porte+journal wiring: `buildAuth`
  (session.Manager + oidc.Kit + local.Kit over portepg), `buildRouter` (httpx chain, health,
  apiref, SPA last), graceful shutdown, journal Wrap on slog.
- **Sablier** `apps/api/internal/env/env.go` — the `Porte()` config shaping + SSO_ONLY.
- **Sablier** `apps/api/modules/` — router/controller/service/types/documentation split.
- **porte SPEC** at `~/.mycelium/memory/conventions/facile-suite-auth.md` + porte repo SPEC.md.
- Suite go.mod floor (Sablier): tronc v0.12+, porte **v0.2.10**, Journal sdk/journal,
  chi, gorm (postgres + sqlite test drivers), godotenv.

## Phase 1 — Auth seam + Workspace (next)

### K1 — Go side
1. `apps/api/go.mod` — add: porte v0.2.10+, gorm (gorm.io/driver/postgres), chi (have),
   goose NOT yet (Phase 6). Run `go get` in `apps/api`.
2. `apps/api/internal/database/database.go` — gorm Open from `appEnv.DatabaseURL` (tronc
   Core). `db.DB()` BEFORE any migration call (the old bug class).
3. `apps/api/internal/auth/users.go` — `auth.NewUserStore(db)` implementing porte's user
   interface (upsert by email, count, find): model the legacy `userTable` (schema:
   `apps/api-legacy/src/database/schema.ts` — CUID2 id, email, name, argon2 hash,
   oidc_subject, roles). porte `local` verifies argon2id at suite params (64MiB/3/2) —
   legacy hashes from better-auth are compatible; the `isLegacyHash` path covers `$2a$`/`$2b$`.
4. `apps/api/main.go` — port Sablier's `buildAuth`: `session.New`, `oidc.New`,
   `local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 8})`.
   Mount `/api/auth/*` (sessions + oidc + local routes) BEFORE the legacy proxy.
5. `apps/api/internal/env` — add the porte env block (OIDC_ISSUER/CLIENT_ID/CLIENT_SECRET/
   REDIRECT_URL/SUCCESS_URL, SSO_ONLY) shaping `porte.Config` exactly like Sablier.
6. **Workspace module** `apps/api/modules/workspace/{router,controller,service,types,
   documentation}.go` — port these routes (path under `/api/workspace`, auth required):
   - `GET /` — list user workspaces (getUserWorkspaces)
   - `GET /{workspaceId}/members` — list members (getWorkspaceMembers)
   - `POST /{workspaceId}/logo` — upload logo (local storage, `UPLOAD_DIR`)
   - `DELETE /{workspaceId}/logo` — delete logo
   Response shapes come from `apps/api-legacy/src/workspace/index.ts` + controllers
   (contract parity, below). Mount BEFORE the proxy; the proxy then skips these.

### K2 — Frontend (parallel, forced by the porte swap)
7. `apps/web/src/lib/auth-client.ts` — replace better-auth client with a thin fetch client
   against porte: `GET /api/auth/config`, `GET /api/auth/oidc`, `POST /api/auth/oidc/callback`,
   `POST /api/auth/local/{sign-in|register}`, `POST /api/auth/logout`.
8. `apps/web/src/components/providers/auth-provider/` (index.tsx, hooks/use-auth.ts) —
   re-plumb to the new client; session = porte cookie, not better-auth session fetch.
9. `apps/web/src/routes/auth/*` (sign-in, sign-up, verify-otp, check-email) — porte flow:
   OIDC redirect or local sign-in; render the `error` query param OUTSIDE any SSO_ONLY gate.
10. Every mutating fetch in the web app: add `X-Facile-CSRF` header (porte reads cookie
    before Authorization; a cookie-authenticated write without the header = 403).
    Test with the browser's transport, not a bearer token.

### Contract parity (every flipped feature)
Capture the legacy OpenAPI (`GET /docs/openapi.json` from a running legacy API, or
`apps/api-legacy/src/utils/openapi-spec.ts`) as a fixture; assert the Go handler answers the
same JSON shape for the routes the frontend actually calls. Route-by-route, flip off the
proxy when parity holds.

## Later phases (from the original plan — order preserved)
- **Phase 2 — Core domain**: project, column, task (incl. bulk/import/export/number), label,
  task-relation, comment, activity. Parity per flip.
- **Phase 3 — Collaboration**: invitation, share-link, workspace-webhook, notification +
  preferences, time-entry, search, workflow-rule.
- **Phase 4 — Events + realtime + scheduler**: `events` (enveloppe envelope, keyed on
  `actor_email`) over **antenne-client/go** — NOT "nook pool"; the suite renamed Nook→Antenne
  2026-08-07, and the `nook-pool`/`nookPool*` names still in `apps/api-legacy` are the SAME
  bus under the old name, scrubbed by the port. WS broadcast (gorilla + Redis pub/sub),
  scheduler (due-date reminders).
- **Phase 5 — Integrations**: github, gitea, slack, discord, telegram, generic-webhook,
  plugins.
- **Phase 6 — Decommission**: storage (S3/local presigned) native; goose baseline from live
  schema; Drizzle auto-migrate OFF; delete `apps/api-legacy`; single container; Dokploy
  deploy; re-open the fresh-vs-preserve DB question for the prod cutover.

## CI (add any time, no dependency)
- `.github/workflows/filet.yml` — mirror `GoSvelteBoilerplate/.github/workflows/filet.yml`
  (FacileStudio/filet@v0.3.0, fail-on error, ANTENNE_FILET_SECRET webhook wiring).

## Exit criteria (per step)
`go build ./...` clean · `filet check .` exit 0 (never adjust thresholds) · `mise run test`
green · parity vs captured OpenAPI where a feature flips · porte login works in the browser,
CSRF enforced · events reach Antenne · single container serves SPA + `/api`.

## Gotchas
- Commit with `git commit --no-verify` (husky runs biome + full pnpm build; commitlint
  wants conventional commits).
- Go toolchain: `mise` (apps/api/mise.toml) or plain `go` 1.26. Run from `apps/api/`.
- filet.yml ignores `apps/api-legacy`, `apps/web`, `apps/docs`, `apps/site` — biome owns the
  TS. filet's `arch.*` findings must never be disabled.
- The Go Docker stage uses `golang:1.26-alpine`; keep go.mod and the image in lockstep.
- Do not `git mv` the package NAME — only directories. The lockfile importer + `link:`
  entries must match the workspace structure or `--frozen-lockfile` fails.
- The legacy runtime in Docker keeps `/app/apps/api` (not api-legacy) + the node_modules
  symlink — do not "simplify" that away.
- Legacy nginx MCP well-known endpoints were dropped with nginx (MCP stays TS/out of scope).

## Skip (YAGNI)
- `packages/mcp`, `packages/email` stay TS (separate components, not the API backend).
- `apps/docs`, `apps/site` untouched.
- No better-auth session-row migration (sessions are short-lived; users re-login).
- No security-header middleware in Go: the suite's Go apps don't ship nginx's headers either
  (Sablier doesn't); dropping them is consistent, not a regression to patch.