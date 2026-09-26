# Opus backend → Go port plan

Cold-start handoff. A fresh session with NO prior conversation can pick up any phase from
this document. Branch `feat/go-backend`. Commits with `git commit --no-verify` (lefthook runs
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

## Status — where the port stands right now
- **Phase 0 DONE** (scaffold + Docker single-container dual-run): commits `1c3b48f7`,
  `7cf0d993`, `15a4e3dd`.
- **Phase 1 K1 DONE** (commit `d733de31`): the Go auth seam (porte over portepg with a GORM
  UserStore, `/auth/me`, local `/auth/register`+`/auth/login` answering `{token,user}`) and
  the workspace module (list, members, logo upload/delete/serve). `go build`/`vet`/`test`
  pass and `filet check .` exits 0 — but exit 0 is not zero findings: both routers carry
  warn-level complexity debt, see "Module architecture" below.
- **Phase 1 K2 DONE** (commits `bcd49d32` Go, `b5101033` workspace/team web,
  `9f7ad953` auth-surface web, `44032b13` keystone): frontend re-plumbed from better-auth
  to the porte fetch client. Contract lived in `docs/k2-contract.md` (now a frozen record).
  Summary: workspace CRUD + membership + full invitation lifecycle Go-native (int64
  `Invitation` table, `pending→accepted` inserts a membership); `PATCH /auth/me` +
  `POST /auth/change-password` (porte **bumped v0.2.10 → v0.3.0** for `Kit.ChangePassword`);
  the legacy proxy authenticates with porte and stamps `X-Opus-User-Id`/`-Email` + an HMAC
  tag (`AUTH_BRIDGE_SECRET`) that the legacy TS API verifies to short-circuit its
  better-auth lookup — the strangler bridge. Web: cut `apiKey`/`device`/OTP/anonymous/
  magic-link/plugin SSO (better-auth-plugin surfaces), porte flow on the auth pages
  (OIDC redirect + local login, `error` query param rendered outside any `SSO_ONLY` gate),
  `X-Facile-CSRF` on every `@opus/libs` mutation, session via `/auth/me`.
- **NEXT ACTION — Phase 2**: the project/column/task block, the biggest module by LOC.
  Before that, browser-transport verification of K2 (cookie + `X-Facile-CSRF`, not a
  bearer token) per the exit criteria; the plan's "porte login works in the browser"
  gate has not been exercised in a live browser yet. Also before Phase 2: refactor the two
  existing routers into the composer pattern (see "Module architecture"). Phase 2 is the
  biggest module and will copy whatever shape it finds, so fix the shape first rather than
  propagating it.

### K2 leftovers / things the next session should know
- `pnpm install` was needed to re-link `@opus/api` into `apps/web/node_modules` after the
  Phase 0 move — `tsc` reported `TS2307` until then. Pre-existing baseline typecheck errors
  (~19 unique) in `apps/api-legacy` (pulled in via `@opus/api` types) and phase-2 test/UI
  fixtures are NOT K2 regressions (0 errors in any K2-owned path).
- `docs/k2-contract.md` is a frozen record of the K2 contract (keystone surface, endpoint
  table, bridge design). Keep for parity checks; it is authoritative for what K2 shipped
  and the endpoint shapes future phases extend.
- Route shadow to remember: Go `GET /invitation/{invitationId}` owns `/api/invitation/*`;
  the legacy `/invitation/pending` and `/invitation/public/:id` are dead/kept — the web
  pending-invites list was re-pointed to `authClient.organization.listUserInvitations()`
  (`apps/web/src/fetchers/invitation/get-pending-invitations.ts`), and the public accept
  page fetcher was rewritten to plain `fetch(getApiUrl("/invitation/public/{id}"))`.
- Settings UI: the Developer/API-keys tab and its components were deleted with the api-key
  surface (`components/settings/api-key-*`, `settings/account/developer.tsx`);
  `routeTree.gen.ts` regenerates on next build.

If this document disagrees with the repo, the repo is the truth and this page is stale—
lint it and push the correction.

## Decisions (baked in — re-open only if something breaks against prod)
- **The Go backend owns its own schema on its own database.** The legacy API still
  derives `DATABASE_URL` from `POSTGRES_*` (the CUID2 `opus` database; Drizzle remains
  the live schema authority for the legacy app). The Go binary is scoped to its own
  `GO_DATABASE_URL` (default `postgresql://…/opus_go`), created by
  `deploy/db/init/01-opus-go.sql` on a fresh volume or `createdb` by hand. This REPLACES
  the original "one shared dev DB, Drizzle owns the schema" rule: it could not survive the
  int64 decision below, and two processes owning one schema is exactly the bug class the
  old rule existed to avoid.
- **Identity is int64, matching the suite and porte.** Opus was the only suite app on
  CUID2 text ids; porte froze `Identity.UserID int64` across its whole contract (sessions,
  identities, stores). The port adopts the suite's int64 model rather than a custom
  int64↔text surrogate store — the original plan step 3 sketched a "CUID2 + argon2 hash +
  oidc_subject + roles" user model that does not match the real schema anyway (better-auth
  keeps the hash in `account`, and there are no `oidc_subject` or roles columns). The
  frontend treats ids as opaque strings, so int64 serialises unchanged through URL
  building. Live prod still runs the old TS API's CUID2 data; a real data migration is a
  Phase 6 cutover concern.
- **Schema ownership of the Go tables is GORM AutoMigrate** (like Sablier), not goose.
  Goose on its own baseline remains a Phase 6 step, from a `pg_dump` of the live DB at
  cutover.
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

## Phase 1 — Auth seam + Workspace (IN PROGRESS)

### K1 — Go side: DONE (this session)
- `go.mod`: porte v0.2.10, gorm + postgres/sqlite drivers.
- `apps/api/schemas/` — Go-owned `users`, `workspaces`, `workspace_members` (int64 PKs),
  GORM AutoMigrate via `schemas.Migrate` in `run()`.
- `apps/api/modules/auth/` — `UserStore` (porte `UserStore` + `PasswordUserStore`,
  int64 ids, kept with the Service like Sablier); Service (RequireAuth, IdentityForUser,
  LoadUser, Register/Login → `{token, user}`); `/auth/me`; local `/auth/register` +
  `/auth/login` (the local kit's own Mount is deliberately NOT used: it would collide
  with these paths and answers porte's lean `{user_id, token}`).
- `apps/api/internal/{authcontext,middleware}` — porte session middleware + context
  hydration.
- `apps/api/internal/env` — porte block (OIDC_*, SSO_ONLY) + `Porte()`, `UploadDir`.
- `apps/api/main.go` — `buildAuth` ported from Sablier (session.Manager + oidc.Kit +
  local.Kit over portepg); router mounts sessions, kit, auth, workspace BEFORE the legacy
  proxy; `schemas.Migrate`; upload dirs created.
- `apps/api/modules/workspace/` — GET / (list), GET /{id}/members, POST/DELETE /{id}/logo,
  GET /{id}/logo (public, matching legacy). Logo files under `UPLOAD_DIR/workspaces/{id}`;
  logo column stamped as cache-buster.
- `apps/api/internal/storage/logo.go` — Save/Delete/Read workspace logos.
- Tests: `modules/auth/users_test.go`, `modules/workspace/service_test.go` (sqlite
  in-memory). `go build`/`vet`/`test`/`filet check .` all green.
- Infra: `deploy/opus-entrypoint.sh` scopes the Go process to `GO_DATABASE_URL`;
  `deploy/db/init/01-opus-go.sql` creates `opus_go`; compose mounts it; `.env.sample`
  documents `GO_DATABASE_URL` / `POSTGRES_DB_GO`.

### K2 — Frontend (next, forced by the porte swap)
7. `apps/web/src/lib/auth-client.ts` — replace better-auth client with a thin fetch client
   against the Go API: `GET /api/auth/config`, `GET /api/auth/oidc`, `GET /api/auth/me`,
   `POST /api/auth/login` and `POST /api/auth/register` (both answering `{token, user}`),
   `POST /api/auth/logout`. No `/auth/local/*` paths exist — porte and this port serve the
   plain `/auth/login` + `/auth/register`.
8. `apps/web/src/components/providers/auth-provider/` (index.tsx, hooks/use-auth.ts) —
   re-plumb to the new client; session = porte cookie, user = `/auth/me`, bearer token kept
   for the fetch wrapper.
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

## Module architecture (the contract every module follows)

Source of truth: the `architecture` block in `filet.yml` at the repo root, inherited verbatim
from the Go-family fork source (Sablier's own `filet.yml`), plus Sablier
`apps/api/modules/projects/` as the reference implementation.

`apps/api/modules/<feature>/` is one responsibility per file:

| File | Owns | Rule |
|---|---|---|
| `router.go` | HTTP only: path, method, middleware, decode, status code | **Required, mechanically enforced** |
| `controller.go` | Request validation and request/response shape conversion | No DB access |
| `service.go` | Business logic and persistence; holds the `*gorm.DB` | The only file touching the ORM |
| `types.go` | Request and response structs | The wire shape |
| `documentation.go` | `var Documentation`, the apiref registry feeding `/docs` | |
| concern files (`invitations.go`, `management.go`, `porte.go`, …) | Allowed when a concern needs its own file | May not be used to skip `router.go` |

Enforcement is one glob:

    architecture:
      requiredDirs: [apps/api]
      requiredFiles:
        "apps/api/modules/*": [router.go]

The fork source spells out why this rule exists: with no architecture block, `filet check
-fail error` can only fail on a file that does not parse, so the CI light is green by
construction and the Antenne alert never fires. Two consequences:

- A package that owns no HTTP does NOT belong in `modules/`. Registre's `keys` and `clients`
  were first written there with a filet.yml exception excusing the missing `router.go`; the
  exception was deleted and they moved to `internal/`. Do not add an exception.
- A module serving a non-chi protocol (WebDAV, CalDAV) overrides the glob to an empty list
  with a named reason. Most-specific pattern wins.

Wiring in `main.go`, two calls per module:

    workspaceService := workspace.NewService(db, appEnv.UploadDir)
    workspace.RegisterRoutes(api, workspaceService, authService)

`NewService` returns the service and stores an **unexported** controller on it
(`service.controller = newController(service)`); route handlers call
`service.controller.method(...)`. `RegisterRoutes(router chi.Router, service *Service,
authService *auth.Service)` is the signature every module shares.

### The composer pattern, and the trap it exists for

filet's `go.func.complexity` and `go.func.statements` descend into anonymous function
literals and charge them to the enclosing top-level function: statements at any depth,
complexity at depth + 1. A chi router god-function with one closure per route therefore
carries every closure's `if err != nil` inside `RegisterRoutes` and can never be made to pass
by reformatting. The closures must move to top-level functions, which resets the budget.

So `RegisterRoutes` becomes a composer: ~5 lines, complexity near 0, calling group helpers
(`registerCredentials`, `registerLogin`, per-resource `register*Routes`), and
`documentation.go` composes `[]documentation.Route` group helpers through nested `append`.
Plume and Vision `apps/api/modules/{auth,clients,verify,documents}` are the worked examples.
One Go dialect quirk: `...` spread is legal only on the final call argument, so
`append(a(), b()..., c()...)` does not compile — nest it.

### Current debt at `a09b3a9d` — `filet check .` exits 0 with 17 warns

`exit 0` is not "clean". The gate is `-fail error` and these are warn level, so CI is green
while the debt sits there. At `a09b3a9d`:

- `apps/api/modules/workspace/router.go:20` — cognitive complexity **160** (limit 15), 295
  lines, 234 statements; the file is 336 lines.
- `apps/api/modules/auth/router.go:26` — cognitive complexity **34** (limit 15), 71 lines, 58
  statements.
- `apps/api/modules/workspace/management.go` — 424 lines; `Update` at complexity 16.
- `apps/api/modules/workspace/invitations.go` — 305 lines.
- `auth/service.go` — `Register` and `ChangePassword` take 6 parameters (limit 5);
  `workspace/service.go` — `Logo` returns 4 values (limit 3).
- `main.go:237` — server implements `io.Closer` with no `Close()` call.

Both routers are the god-function shape Sablier still ships, so they were copied faithfully
from the reference; the reference is simply behind the composers Plume and Vision landed in
September. Expected noise, not defects, and do NOT "fix" per module: `go.global.mutable`
fires on every `var Documentation` (three here), and `go.comment.inbody` fires on the info
tier. Sablier's `modules/` carries the same findings at a larger count.

## Reference implementation (read before writing Phase 1 code)
- **Sablier** `apps/api/main.go` — the canonical tronc+porte+journal wiring: `buildAuth`
  (session.Manager + oidc.Kit + local.Kit over portepg), `buildRouter` (httpx chain, health,
  apiref, SPA last), graceful shutdown, journal Wrap on slog.
- **Sablier** `apps/api/internal/env/env.go` — the `Porte()` config shaping + SSO_ONLY.
- **Sablier** `apps/api/modules/` — router/controller/service/types/documentation split.
  Copy the file split, NOT its routers: Sablier's are the pre-composer god-functions.
- **Plume / Vision** `apps/api/modules/{auth,clients,verify,documents}` — the composer
  pattern both `router.go` and `documentation.go` should follow (see "Module architecture").
- **porte SPEC** at `~/.mycelium/memory/conventions/facile-suite-auth.md` + porte repo SPEC.md.
- Suite go.mod floor (Sablier): tronc v0.12+, porte **v0.2.10**, Journal sdk/journal,
  chi, gorm (postgres + sqlite test drivers), godotenv.

## Later phases (from the original plan — order preserved)
- **Phase 2 — Core domain**: project, column, task (incl. bulk/import/export/number), label,
  task-relation, comment, activity. Parity per flip.
- **Phase 3 — Collaboration**: invitation, share-link, notification +
  preferences, time-entry, search, workflow-rule.
- **Phase 4 — Events + realtime + scheduler**: `events` (enveloppe envelope, keyed on
  `actor_email`) over **antenne-client/go** — NOT "nook pool"; the suite renamed Nook→Antenne
  2026-08-07, and the `nook-pool`/`nookPool*` names still in `apps/api-legacy` are the SAME
  bus under the old name, scrubbed by the port. WS broadcast (gorilla + Redis pub/sub),
  scheduler (due-date reminders). **Antenne is the only outbound event path** (2026-09-26):
  both hand-rolled webhook modules are dropped, see below.
- **Dropped integrations (2026-08-12)**: github, gitea, slack, discord, telegram,
  **generic-webhook**, plugins and oauth — none of it gets ported, none of it is Facile
  style: extra integration surface on a product whose own rules say "features solve real
  problems, nothing more". There is no Phase 5 anymore. They die at cutover with
  `apps/api-legacy`; the frontend feature code (fetchers/hooks for github-integration,
  slack-integration, generic-webhook-integration, …) is scheduled for removal when the
  legacy API is deleted in Phase 6. `oauth` is redundant anyway: it was better-auth's
  `/auth/id-token` getter, and porte's OIDC kit owns identity tokens now. Watch:
  `external-link` (task external links, the old GitHub/Gitea binding) never had a phase;
  with the integrations gone, port it only if the UI keeps the generic resource link.
- **Dropped webhook modules (2026-09-26)**: **`workspace-webhook`** joins the drop list
  (it was scheduled for Phase 3) alongside `generic-webhook`. Neither is connected to
  Antenne: both subscribe to the local in-process bus (`subscribeToEvent`) and POST JSON to
  a user-supplied URL with an HMAC-SHA256 `X-Opus-Signature`, re-implementing outbound
  delivery the bus already owns. Only `nook-pool` (`@facile/antenne`, `app: "Opus"`,
  `instance_id: workspaceId`, per-workspace `nook_pool_config` + outbox + processed-event
  tables) speaks to Antenne, and Antenne already routes to delivery targets. Three modules
  collapsing to one is the "features solve real problems, nothing more" rule applied.
  Frontend surface to delete with them: the webhook section of
  `apps/web/src/routes/_layout/_authenticated/dashboard/settings/workspace/notifications.tsx`,
  `apps/web/src/components/project/generic-webhook-integration-settings.tsx`, and the
  `fetchers|hooks/{workspace-webhook,generic-webhook-integration}/` trees (~1450 LOC of
  legacy TS that no longer needs a port). Consequence to keep in mind: Opus users lose the
  ability to point events at their own arbitrary endpoint; targets are registered in
  Antenne, not in Opus. Acceptable if Antenne is the suite's delivery path, re-open if
  per-project outbound URLs turn out to be a product requirement.
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

Two caveats on that first line. `filet check .` exit 0 does not mean zero findings — the gate
is `-fail error`, so warns pass silently; read the output, do not just read the code. And
`mise run test` currently fails on the installed mise (2026.6.14), see Gotchas; the underlying
`go test ./...` is the real gate until that is fixed.

## Gotchas
- Commit with `git commit --no-verify` (lefthook runs biome + full pnpm build on pre-commit;
  commit-msg wants conventional commits and derives the gitmoji).
- `mise run test` fails with the currently installed mise (2026.6.14): it errors on the
  Taskfile's `depends=[":task"]` references — needs mise ≥ 2026.8.5 or a depends rewrite.
  The underlying gates (`go test`, `go vet`, `filet check .`) all pass directly.
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