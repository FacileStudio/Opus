# Changelog

All notable changes to Opus are documented here.

## [Unreleased]

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
