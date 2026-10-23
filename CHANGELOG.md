# Changelog

All notable changes to Opus are documented here.

## [Unreleased]

## [2.13.0] - 2026-10-23

### Added

- Task cards and list rows show subtask progress (done of total).
- Short task links: `/dashboard/workspace/{id}/task/SLUG-N`. "Copy task link" now copies this form.
- Notifications can be marked unread and deleted one at a time.

### Changed

- Cards, columns, panels, menus, popovers and dialogs have no borders.
- The sidebar is a floating card, shows your name in the footer, and no longer shows the version.
- The Gantt view opens centered on today.

## [2.12.0] - 2026-10-23

### Changed

- The web app uses the muse theme: chroma-zero light and dark tokens.
- Cards, board columns and task cards no longer have borders. Fills separate them.

### Fixed

- The task sidebar no longer requests `GET /api/task` with an empty id, which
  returned a 404.
- The seed page declares its language, so `biome ci` passes.

## [2.11.2] - 2026-10-23

### Fixed

- Board cards show a pointer on hover and a grabbing cursor while pressed or
  dragged, instead of the move cursor.
- Clearing a task's priority sends `no-priority`, which the API accepts, instead of
  an empty string, which it rejects with a 400.

## [2.11.1] - 2026-10-23

### Fixed

- Task cards on the board did not open and could not be dragged between columns.
  The Go API returns numeric ids, the web app compares and routes on string ids,
  so `?taskId=` was dropped and every drag lookup missed. The API client now turns
  numeric `id`, `*Id` and `*Ids` fields in JSON responses into strings.

## [2.11.0] - 2026-10-13

### Added

- Personal API tokens: `POST /auth/api-tokens` mints one, `GET /auth/api-tokens`
  lists the caller's own, `DELETE /auth/api-tokens/{id}` revokes one. A token is a
  labelled porte session, so it authenticates as its owner on every route a browser
  login does, and it never idles out — which is what a CLI or a cron job needs. The
  plaintext comes back once, on create; the server stores only the hash.
- The account settings area has a Developer page that lists those tokens, creates
  one with the value and the `~/.opus.yml` / `OPUS_TOKEN` snippets shown once, and
  revokes one behind a confirmation.

## [2.10.0] - 2026-10-13

### Changed

- The settings area is wider and centred, with a refreshed sidebar, navigation
  and workspace switcher.

### Added

- Workspace members settings page.

### Fixed

- Creating a task no longer answers 400. The web create form sent an empty
  `userId` for an unassigned task, which the strict request decoder rejected as
  an invalid id; an absent assignee is now omitted. The same empty-string leak
  is fixed on the full task update, the actor update and the due-date update, so
  unassigning a task and clearing its due date work again.
- `PATCH /task/bulk` understands the `addLabel` operation the board and backlog
  toolbars send, applying a workspace label to every selected task.

## [2.9.2] - 2026-10-12

### Fixed

- Request bodies carrying ids accept the decimal string form the web client and the MCP server
  actually send. Every create/update endpoint binding an id to `int64` — project, label, task
  (create, update, actor, move, bulk), activity comment, time entry, share link, task relation,
  workflow rule and column reorder — answered `400 invalid JSON body` when the id arrived as a
  string. The wire type now decodes both forms strictly (`internal/wireid`), with tests pinning
  the exact payload the frontend sends.

## [2.9.1] - 2026-10-12

### Fixed

- `GET /auth/me` answers the bare profile while the client expected `{token, user}`, so every
  session probe threw after OIDC sign-in and an authenticated browser sat on the landing page
  instead of entering the app.
- A failed OIDC login redirected to `/login`, a route the SPA did not serve. It now forwards the
  reason to the sign-in page, which displays it.

## [2.9.0] - 2026-10-12

### Added

- Landing page at `/` with the feature overview and sign-in CTA. Visitors who already hold a
  session pass straight through to the dashboard.
- Split-screen auth shell (brand panel plus form) shared by sign-in and sign-up.

### Fixed

- Auth, session and websocket requests doubled the `/api` prefix (`/api/api/...`) whenever
  `VITE_API_URL` already ended in `/api`: the SSO button 404'd, `GET /auth/me` never returned a
  session, and the project websocket never connected. All API URL building now goes through one
  normalizer, covered by regression tests. Completing OIDC sign-in also requires
  `https://opus.facile.studio/api/auth/oidc/callback` registered as an Opus redirect URI in
  Registre's seed.

## [2.8.2] - 2026-10-12

### Fixed

- Healthcheck probe now uses the native `/usr/local/bin/opus-api healthcheck` command instead of GNU `wget --spider`.
- Mount `HEAD` handlers on `/health` and `/api/health` to answer HTTP HEAD probes with 200.

## [2.8.1] - 2026-10-02

### Changed

- The container listens on port `80` again, reverting the `27538` default introduced in 2.8.0.

## [2.8.0] - 2026-10-02

### Changed

- The single container now listens on port `27538` instead of `80`. Point your reverse proxy
  (Traefik/Dokploy) at container port `27538`.

### Fixed

- Helm chart `opus.service` port, Ingress and HTTPRoute backends now default to the port the image
  actually listens on, instead of the decommissioned `5173`.
- The deployment container's `PORT` is now derived from `opus.service.targetPort`, so the service
  target and the listening port stay in sync.
