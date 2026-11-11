# Sablier sync

Opus and Sablier keep projects and tasks in step over the Antenne pool, and Opus mirrors Sablier's
time entries. Each app publishes its local changes and applies the other app's. Code:
`apps/api/internal/poolsync` (inbound), `apps/api/internal/events` (outbound). The wire contract is
v2 (2026-11-03): every field below is additive, an absent field never clears anything, and an
explicitly present empty value does.

## Identity

Every project and task has a `facile_id`, a unique string minted by the app that creates the row.
Opus mints 26 base32 characters (crypto/rand). Events always carry it, and an inbound event matches
rows by `facile_id` only, across all workspaces.

Migration `00002_facile_id_and_pool_ledger.sql` adds the columns, keeps every existing `facile_id`
and adds unique indexes. Rows with a blank id get a random one. When several rows share an id, which
the per-workspace and per-project unique keys of the old TypeScript schema allowed, the oldest row
(lowest `id`) keeps it and the others get a random one. The migration can run again on a
half-migrated database. A failed migration stops the API with exit code 1.

## Channels

| Direction | Channels |
|---|---|
| Opus publishes | `project.{created,updated,deleted}`, `task.{created,updated,deleted}`, plus `task.status_changed`, `task.actor_changed`, `comment.created`, `workspace.created` |
| Opus applies | `project.{created,updated,deleted}`, `task.{created,updated,deleted}`, `time_entry.{created,updated,deleted}` |

Creating, renaming, re-iconing, describing or deleting a project publishes a project event.
Creating, editing, moving, importing, bulk-updating or deleting a task publishes a task event.
`actor_email` on a task event is the task's assignee, not the user who made the edit. Archiving a
project publishes nothing. Deleting a project publishes `task.deleted` for each of its tasks, then
`project.deleted`. A change applied from Sablier is never published, so the Agenda does not see
assignee or due date changes made in Sablier.

## Outbound delivery

Every project and task event goes into the `antenne_outbox` table first, through antenne-client's
`outbox` package (table from migration `00005_sablier_sync_v2.sql`, converted to the library's
shape by `00009_antenne_outbox_v1.sql`). The library's drainer sends unsent rows in id order every
2 seconds while the pool is connected, marks each one sent, and stops the pass at the first
failure, so a dead pool delays events but never drops or reorders them. After a failure it backs
off up to 5 seconds. Sent rows are pruned after 7 days; an erasure deletes the sent rows that name
the erased person at once. Idempotency keys read `<channel>:<facile_id>:<unixnano>`. The payload
column is `jsonb`, so the sent bytes carry Postgres's key order, not Go's: same JSON, different
bytes.

A project event carries `space`: `registre_space_id` for a workspace linked to Registre, or
`personal: true` with the owner's porte subject and lowercased email for a personal workspace.
Projects of any other workspace, and their tasks, are never published. Events carry `updated_at`
(RFC3339Nano). Task events carry `description`, `due_date` (`YYYY-MM-DD` or `null`), `priority`
(`no-priority` goes out as `none`) and `actor_subject`. A row without a `facile_id` gets one minted
before it is published.

## Reconcile

`POST /api/sync/reconcile` (any workspace owner or admin, answers 202) and every boot, once the
first Registre snapshot is applied (right away without OIDC), assign missing facile_ids and queue
`project.created` for every active syncable project, then `task.created` for each of their tasks.
Archived projects are left out. The receiver treats a `created` for a known id as an update.

## Inbound rules

1. Only events whose `app` is `Sablier` are applied. Events from `Opus` (its own echo), any other
   app, or with no envelope are ignored.
2. An event with no `timestamp`, or one older than 35 days (the ledger window), is dropped, so a
   pool replay of old history cannot recreate rows or revert newer edits. An event whose
   `facile_id` is all digits is dropped as a legacy numeric id. A task update older than the task's
   `updated_at` is ignored.
3. Each event's `idempotency_key` goes into the `pool_processed_events` table in the same
   transaction as the change. A key seen before is skipped. Rows older than 35 days are pruned at
   startup.
4. Opus writes inbound changes straight to the database and never calls the publisher, so an
   applied change is not published back.
5. A task whose `project_facile_id` matches no local project is dropped with a warning. The event
   is not recorded in the ledger, so a redelivery after the project arrives applies.
6. `created` for a known row is an update. `updated` for an unknown project or task is ignored.
   `deleted` for an unknown row is a no-op. Deleting a project deletes its tasks.
7. Project and task updates are last-writer-wins: an update whose `updated_at` (else the envelope
   timestamp) is older than the row's `updated_at` is ignored.
8. Task `description`, `priority` and `due_date` apply only when present. `normal` reads as
   `medium`, `critical` as `urgent`, `none` as `no-priority`. A different `project_facile_id` moves
   the task within its workspace; a move across workspaces is ignored, as in the Opus UI.
9. Payloads are read from the envelope (`{"app", "timestamp", "payload": {...}}`).
10. The assignee is matched by `actor_subject` (porte identity under `OIDC_ISSUER`), then by
    `actor_email` ignoring case, and must be a member of the task's workspace. No match leaves the
    assignee unchanged.

## Twin merge

A `created` with an unknown `facile_id` first looks for a twin: a project in the same resolved
workspace, or a task in the same project, with the same name ignoring case and a different
`facile_id`. A twin counts only while unlinked (`synced_at` is null). A row is linked once an
inbound event created or updated it, or once a user action published it; reconcile does not link.
Both sides keep the smaller id: the twin adopts the incoming id only when it sorts first byte-wise,
otherwise the event is dropped and Sablier adopts ours. After a project adopts an id, Opus queues
`task.created` for all its tasks so Sablier learns them under the new project id.

## Time entries

Time is one-way, Sablier to Opus. A `time_entry` event upserts a row of `time_entries` keyed by
its `facile_id`, with `source = 'sablier'`, on the task named by `task_facile_id` and the user
matched by subject, then email. A running entry has no end and a zero duration; a stopped one counts
`paused_duration_ms` out. An entry whose task Opus does not hold yet is skipped and not recorded,
so a redelivery applies. `deleted` removes the mirror. The API answers 409 to editing or deleting a
mirrored entry and returns `source` on every entry. Opus-native entries never leave Opus.

## Target workspace

An inbound project lands in the workspace its `space` resolves to: the one whose
`registre_space_id` matches, or for a personal project the owner's oldest personal workspace (owner
by porte subject, then email). A space that resolves to nothing skips the event with a warning and
leaves it out of the ledger, so a replay or reconcile applies it later. A changed space on
`project.updated` moves the project and its tasks. `ANTENNE_SYNC_WORKSPACE_ID` is only the fallback
for old Sablier payloads that carry no `space`.

## Status mapping

Opus statuses are the slugs of a project's workflow columns (default: `to-do`, `in-progress`,
`in-review`, `done`). Sablier has exactly those four. Opus publishes its own status unchanged. Inbound status values are lower-cased and trimmed.

For an inbound status, Opus picks the first rule that matches:

| Rule | Result |
|---|---|
| Empty | No change (new tasks start in the first open column) |
| A column slug equals the value | That column |
| Value is `done` | The first column marked final |
| Value is `to-do`, `in-progress` or `in-review` | The 1st, 2nd or 3rd open column by position, or the last open column when fewer exist |
| Anything else | Unchanged on update, first open column on create |

A project with no final column keeps a task's status when Sablier sends `done`.

An update does not move a task whose current column already stands for the value Sablier sent.
Without that, renaming a task in Sablier would pull a task in a custom column such as `qa` back to
the first column. A local status stands for `done` when it is `archived` or in a final column, for
`to-do` when it is `planned`, and otherwise for the rank of its open column.

Sablier collapses statuses it does not know to `to-do` on its side, so a custom Opus column reaches
Sablier as `to-do` unless Sablier maps it.

## Failure behavior

The pool is optional at every step. Unset `ANTENNE_URL` or `ANTENNE_SECRET` turns publishing and
listening off. A dead pool does not block a local write: the publisher dials in the background with
backoff, and the client reconnects on its own. Handler errors are logged and the event is acked.
