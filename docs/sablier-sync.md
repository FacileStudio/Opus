# Sablier sync

Opus and Sablier keep projects and tasks in step over the Antenne pool. Each app publishes its
local changes and applies the other app's. Code: `apps/api/internal/poolsync` (inbound),
`apps/api/internal/events` (outbound).

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
| Opus applies | `project.{created,updated,deleted}`, `task.{created,updated,deleted}` |

Creating, renaming, re-iconing, describing or deleting a project publishes a project event.
Creating, editing, moving, importing, bulk-updating or deleting a task publishes a task event.
`actor_email` on a task event is the task's assignee, not the user who made the edit. Archiving a
project publishes nothing. Deleting a project publishes `task.deleted` for each of its tasks, then
`project.deleted`. A change applied from Sablier is never published, so the Agenda does not see
assignee or due date changes made in Sablier.

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
5. `created` for a known row is a no-op. `updated` for an unknown row creates it. `deleted` for an
   unknown row is a no-op.
7. Payloads are read from the envelope (`{"app", "timestamp", "payload": {...}}`).
8. `actor_email` is matched to a member of the task's workspace, case-insensitively. No match
   leaves the assignee unchanged.

## Target workspace

Opus has no per-workspace pool config. `ANTENNE_SYNC_WORKSPACE_ID` names the workspace that
receives projects created in Sablier. Unset, such projects are skipped with a warning at startup.
Projects Opus created publish from every workspace, and their updates, deletes and tasks apply
wherever they live.

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
