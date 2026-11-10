# Changelog

All notable changes to Opus are documented here.

## [2.34.0] - 2026-11-10

### Changed

- The licence changes from MIT to AGPL-3.0-only. Releases up to and including v2.33.0 remain under MIT; the next release is the first under AGPL-3.0-only. The Kaneo-derived parts keep their MIT notice in `LICENSE-MIT`, and `NOTICE` records the derivation. A commercial licence is available from Facile Studio.

### Fixed

- The MCP confirmation before deleting a project no longer says "its 1 tasks": it counts the tasks and drops the count when there are none.
- The unused `apps/web/nginx.conf` and `nginx.opus.conf`, which still routed the old `/api/mcp` metadata path, are removed: the Go binary serves the SPA and the MCP metadata itself.

## [2.33.0] - 2026-11-10

### Added

- A remote MCP server at `/mcp` (Streamable HTTP, stateless, spec 2026-07-28) on tronc/mcp v0.20.0, behind Registre tokens bound to `MCP_RESOURCE` with the scopes `opus.read` and `opus.write`. Eleven tools: list workspaces, list projects, search tasks, read a task with its comments, create and update a task (status, assignee, dates, labels, project), comment on a task, create and update a project (including archive), and delete a task or a project. Tools go through the same module calls as the HTTP routes, so they see exactly what the person sees and their writes broadcast and sync like the web app's. A first MCP call creates the Opus account (porte/mcpauth v0.2.0 provisioning). MCP is free on every plan.
- Deleting a task or a project asks the person first: an MRTR confirmation in clients that can elicit, otherwise a pending action and a link to the new `/pending` page, where they confirm or dismiss it (fr, en, es, de). Pending actions live in `mcp_actions` (migration 00008).

### Fixed

- The Ideas page spun forever, and its sidebar entry stayed hidden, for a person who never opened a workspace. Both now fall back to the personal scope once the workspace list is known.
- The Ideas plan store forgot the plans on every restart and, until Registre answered, let every request through. It now saves the plans in `forfait_state` (migration 00008) and restores them at boot, and AI stays off until entitlements were received once.
- The docs advertised a stdio MCP package and an `/api/mcp` endpoint that the Go API never served. The README and `docs/` now describe the real `/mcp`.

### Changed

- tronc v0.20.0, porte v0.16.0 and porte/mcpauth v0.2.0, with go-sdk v1.8.0.

## [2.32.0] - 2026-11-10

### Added

- An Ideas page and sidebar entry (lightbulb icon, first in the sidebar, badge with the count of new ideas). Bulle generates ideas for Opus, runs them in place with approval cards, and links to the result and to the conversation in Bulle. Free plans see an upsell with the upgrade link. Available in fr, en, es and de.
- `/api/ideas` routes in the API, proxied to Bulle with a Registre token exchange (audience `bulle`). They need `BULLE_URL`; without it the routes are not mounted. The active workspace's Registre space is passed through, and a plan without `ai.access` gets 402 or `ai_access: false`.

### Changed

- porte v0.15.1 and tronc v0.18.0.

## [2.31.1] - 2026-11-10

### Added

- An MIT `LICENSE` at the repository root.

### Removed

- The unused marketing site (`apps/site`), documentation site (`apps/docs`), local MCP package (`packages/mcp`) and email package (`packages/email`). None was built into the image or deployed. The API's own `/api/mcp` endpoint is unchanged.

## [2.31.0] - 2026-11-09

### Added

- Antenne Pool bus consumer for `git.pull_request` and `git.push` channels, enabling single-funnel GitHub webhook routing through Antenne.

## [2.30.0] - 2026-11-09

### Added

- Native GitHub integration with webhook receiver at `/api/integrations/github/webhook`.
- Pull request and branch linking to tasks based on `{slug}-{number}` convention.
- Automated task workflow state transitions on PR draft/open (`in-progress`), ready for review (`in-review`), and merge (`done`).
- Automatic user attribution on PR actions using Registre's `connections.github` identity claim.
- Linked pull requests display in task sidebar with real-time status badges.

## [2.29.3] - 2026-11-07

### Changed

- `GET /version` serves the version from the root `package.json`, baked into the binary at build time by
  `scripts/version.sh`. The `APP_VERSION` build arg and its Dokploy env key no longer play any part, so the
  endpoint can no longer lag behind the deployed release.

## [2.29.2] - 2026-11-07

### Fixed

- The user menu that opens from the sidebar user card shows a Solar icon beside every option (account,
  settings, language, theme, help, sign out) and on the back row of its submenus. Submenus point right with
  an arrow icon instead of a text chevron.

## [2.29.1] - 2026-11-07

### Fixed

- The project icon picker in project settings has padding again, so the search field is no longer cut at the
  edges. Both pickers (new project and settings) share one grid of fixed-size cells that fills the popover
  width, so icons no longer overlap, including on touch screens.

## [2.29.0] - 2026-11-07

### Added

- Task responses carry `facile_id`, the suite-wide id Sablier keys its task mirror on. Bulle reads it from
  the Opus tools so the time an agent spends on a task lands on that task in Sablier.

## [2.28.1] - 2026-11-07

### Security

- Registre tokens are sealed at rest. porte to v0.13.0: `porte_identities.access_token` and `refresh_token`
  are encrypted with AES-256-GCM under `PORTE_TOKEN_KEY`. Existing rows are sealed once, in the background,
  after boot, and the count is logged. An invalid `PORTE_TOKEN_KEY` stops the boot. With the key unset nothing changes.
- Losing `PORTE_TOKEN_KEY` makes the stored tokens unreadable: people sign in again, nothing else is lost.
  Do not roll back to a build older than porte v0.13.0 after the first reseal.

## [2.28.0] - 2026-11-07
### Changed

- The sidebar footer is the Facile user card with a borderless menu: Facile account, Settings, Language, Theme (system, light, dark), Help and Sign out. It replaces the avatar dropdown and the theme switch.
- Settings use the Facile layout: a section list on the left from 64rem, a drill-in list with a back link on a phone. Order: Profile, Preferences, Notifications, Workspace, Members, Workspace notifications, Nook Pool, Projects, API keys. The section routes and the API key and CLI token calls are unchanged.
- The muse colour, radius and text tokens are now available app-wide (`styles/muse-tokens.css`), not only on the sign-in screen.

## [2.27.0] - 2026-11-07

### Added

- GDPR deletion receiver. Opus polls Registre's deletion list with its OIDC client credentials and accepts Registre's signed push on `POST /api/registre/deletions`. An erased person loses their account, personal workspace (projects, tasks, files), memberships, notifications, tokens and invitations sent to their address. What they wrote in shared workspaces stays without an author (comments, share links, time entries, activity), shown as a deleted user. The objects of the personal workspace (uploads and attachments, on local disk or S3) are deleted before the rows, so a storage failure leaves the account in place and the retry finishes the job; a missing object counts as deleted. Replays are no-ops; the cursor lives in `directory_deletion_state`.
- `GET /version`, from the `APP_VERSION` build arg.
- The login page matches the suite's auth screen: the "Sign in with Facile" pill with the canonical label, Registre sign-up and legal links, and a language switch. The landing page and the sign-in screens are translated in fr, en, es and de (es and de need native review).
- `AssetStorage.Delete`, a batched, idempotent removal of stored objects (S3 `DELETE` or local file).
- CI: API (vet, race tests against Postgres), web (i18n, typecheck, biome, tests, build) and filet.

### Changed

- porte v0.7.0 to v0.12.1, `porte/directory` v0.1.0 to v0.3.0, tronc v0.13.0 to v0.17.1. tronc's security headers are on: enforced anti-framing and a report-only CSP that allows the client's inline scripts, YouTube embeds and the S3 origin.
- Sign-in links an existing account by email only when Registre verified the address, and refuses when the account already holds another subject of the same issuer.
- The Registre `locale` claim is stored on the account at every sign-in.
- Migration `00006_launch_floor.sql`: `porte_login_codes.provider` and `.subject`, `directory_deletion_state`, and `comments.user_id` and `share_links.created_by` become nullable with `ON DELETE SET NULL`.

### Removed

- Local sign-up and password sign-in: `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/change-password`, the sign-up page and the password form. Clients that signed in with the account password must use an API token (`/api/auth/api-tokens`) or `opus login`.
- The `SSO_ONLY` setting. With an issuer set, Opus never mounts password routes.

## [2.26.0] - 2026-11-03

### Added

- `OIDC_CLI_AUDIENCE` mounts porte's device exchange at `POST /api/auth/oidc/device/exchange`, so suite CLIs sign in with the browser approval link from Registre's device grant, over SSH too. Set it to `facile-cli`; unset, the route stays absent and answers 404. It is independent of `OIDC_MACHINE_AUDIENCE`, which stays Opus's own client id.

## [2.25.0] - 2026-11-03

### Added

- Sablier sync v2: projects follow their space both ways (Registre space id, or the owner's personal workspace), tasks carry description, due date and priority, and Sablier time entries on synced tasks are mirrored into Opus as read-only entries (409 on edit or delete).
- A durable outbox for pool events, sent in order and retried, so nothing is lost while the pool is unreachable.
- Reconcile at boot and on `POST /api/sync/reconcile`: same-named projects and tasks in the same space merge onto the smaller facile id, everything else is created on the other side.

### Changed

- `ANTENNE_SYNC_WORKSPACE_ID` is only a fallback for events that carry no space.
- Projects use last-writer-wins on inbound updates, like tasks.

## [2.24.1] - 2026-11-03

### Changed

- Spaces use the share-circle icon.

## [2.24.0] - 2026-11-03

### Added

- Every user has a personal workspace, created for existing users by migration 00004 and for a new account on its first workspace listing, so nobody has to create one first. It is listed first.

### Changed

- Personal workspaces are never imported into Registre or linked to a shared space by name, and cannot be shared or deleted.

## [2.23.0] - 2026-11-03

### Added

- Registre owns workspaces. When OIDC is configured, Opus imports its workspaces into Registre once, retrying with backoff until Registre accepts, then keeps them equal to Registre's shared spaces: at boot, every 5 minutes, and about 2 seconds after Registre posts to `POST /api/registre/sync`. A space links to the workspace with the same name, ignoring case, or creates one.
- People added in Registre are provisioned in Opus before their first sign-in, with their name and picture. People removed in Registre lose the workspace. Local password accounts are never removed by a sync.
- Migration `00003` adds `workspaces.registre_space_id`.

### Changed

- A workspace linked to Registre shows its members read-only with a "Manage in Registre" link. The API refuses rename, delete, invitations and member changes on it with 409 and the Registre link.
- Once Registre owns workspaces, creating a workspace in Opus points to Registre, and `POST /api/workspace` answers 409. `GET /api/config` reports `registreSpacesUrl`.
- Pending invitations to linked workspaces are canceled. Add people in Registre instead.

## [2.22.0] - 2026-11-03

### Added

- My tasks has a columns view that groups your tasks by status, next to the list view. The header toggle remembers your choice.

### Changed

- My tasks shows a summary line with the open, overdue and due-this-week counts, and a loading skeleton while it fetches.
- The list view sorts each project's tasks by due date, then priority, and colors overdue and due-today dates. Rows are links, so a task opens in a new tab with a middle click.

## [2.21.0] - 2026-11-03

### Added

- Opus stores the Registre profile picture URL (the `picture` claim) on every sign-in and profile sync, so the user menu, assignees, comment and activity authors, workspace members and time entries show the picture set in Registre, with initials when there is none. A change in Registre shows on the next page load.
- At boot, Opus backfills the picture URL (`<OIDC_ISSUER>/avatars/<subject>`) for every account with a Registre identity, so users see their picture without signing in again. It also replaces stale image URLs left by earlier providers.
## [2.20.0] - 2026-11-02

### Added

- The API accepts short-lived bearer JWTs minted by Registre for another suite service, such as Bulle acting for a user. Set `OIDC_MACHINE_AUDIENCE` to Opus's client id and `OIDC_SCOPE_PREFIX` to `opus`: a token then needs `opus.read` for GET, HEAD and OPTIONS and `opus.write` for any other method. Opaque API tokens and cookie sessions are unchanged.
- The request identity carries the acting service (`Identity.Actor`) when a delegated token is used.

### Changed

- porte is bumped to v0.7.0.

## [2.19.3] - 2026-10-23

### Fixed

- Task events carry `owner_email`, the user who made the change, so Agenda can show the due date of an unassigned task in that user's calendar.

## [2.19.2] - 2026-10-23

### Fixed

- The floating top bar and the project toolbars have no border, and they now share the same width.

## [2.19.1] - 2026-10-23

### Fixed

- Task events carry the assignee's email in `actor_email`, not the email of the user who edited the task, so Agenda keeps due dates in the assignee's calendar.
- The Sablier sync ignores events from apps other than Sablier, ignores replayed events with legacy ids, and never creates a project or task from an id the apps did not mint.
- Migration `00002` blanks empty and duplicate `facile_id` values before it creates the unique indexes, so it cannot fail on legacy data.
- A failed database migration now stops the boot. The GORM schema fallback is removed; goose is the only migration path.

## [2.19.0] - 2026-10-23

### Changed

- The top bar is a floating pill on every page. The board, backlog and Gantt toolbars float the same way.
- The sidebar follows the Facile suite layout: a borderless card with a 12px radius on the page background, and the content sits flat beside it.

## [2.18.0] - 2026-10-23

### Added

- Projects and tasks sync with Sablier in both directions over the Antenne pool. Create, update and delete apply on the other side, replays are ignored, and Opus never re-publishes a change it received. Set `ANTENNE_SYNC_WORKSPACE_ID` to choose the workspace that receives projects created in Sablier. See `docs/sablier-sync.md`.
- Projects publish `project.created`, `project.updated` and `project.deleted` on the Antenne bus, and moving a task publishes `task.updated`.

### Changed

- Projects and tasks carry a stable `facile_id`, minted at creation and backfilled for existing rows by migration `00002`. Events use it instead of the numeric id or slug.
- Antenne events for projects now always include the description, so clearing it reaches Sablier.

## [2.17.0] - 2026-10-23

### Added

- Tasks publish `task.created`, `task.updated` and `task.deleted` on the Antenne bus, with a `due_date` field (RFC3339 UTC, omitted when unset) so the Agenda can show due dates as calendar events.

## [2.16.0] - 2026-10-23

### Added

- My tasks refreshes live when someone else changes a task in the workspace.

### Changed

- Project events also reach clients listening on the project's workspace.
- The project view switcher and the backlog filter button no longer have borders.

## [2.15.0] - 2026-10-23

### Added

- Live updates: task, comment, label and relation changes reach other open windows without a reload. The window that made the change is skipped.

### Changed

- WebSocket connections require membership in the project's workspace.
- My tasks leaves out tasks from archived projects.
- My tasks, short task links and notification actions are translated in German, Spanish, Dutch, Russian, Ukrainian, Greek and Macedonian.

## [2.14.1] - 2026-10-23

### Fixed

- Pages refetch stale data when you return to them, so My tasks and other lists update after you change a task without a reload.
- A hovered board card has its own fill, different from the column. Cards are lifted in dark mode.

## [2.14.0] - 2026-10-23

### Added

- My tasks page and `GET /api/task/me`: tasks assigned to you, grouped by project.

### Changed

- Cards, popups and the page shell use shadows and fills for separation now that borders are gone.
- The i18n check requires the plural forms each locale needs.
- The web client is typed against the Go API, so the web typecheck passes.

### Fixed

- WebSocket connections no longer fail with a 500 behind the request logger.
- Shared project links group tasks by status, like the board, and no longer crash.

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
