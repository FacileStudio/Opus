# Erasure

When Registre erases a person, Opus deletes what belongs to them and keeps what they wrote for
others, without their name on it.

## How a notice arrives

`workspace.Deletions` wraps porte/directory's receiver. It polls Registre's deletion list with
the OIDC client credentials and accepts Registre's signed push on `POST /api/registre/deletions`.
Both paths call the user store's `EraseAccount` (porte/kit), then porte's own identity eraser.
The notice matches the issuer with or without a trailing slash. A notice for a person Opus never
saw, or one already erased, is a success, so a replay changes nothing.

Registre holds a self-deletion for 30 days before it notifies apps.

## The rule

| Data | What happens |
|---|---|
| The `users` row: email, name, picture, locale, GitHub id | Deleted |
| Identities, sessions, API tokens, login codes | Deleted with the row (`ON DELETE CASCADE`) |
| The personal workspace, its projects, tasks, labels and files | Deleted. Stored objects go first, on disk or S3 |
| Memberships, notifications, notification preferences and rules | Deleted with the row |
| Invitations they sent, and invitations sent to their address | Deleted |
| Pending MCP actions (`mcp_actions`) | Deleted |
| Bus events naming their email or subject (`antenne_outbox`), sent or still queued | Deleted. A queued event is never delivered; the next reconcile re-sends shared projects and tasks |
| Comments, activity entries, time entries, assets and share links in shared workspaces | Kept. The author becomes empty (`ON DELETE SET NULL`); the client shows a deleted user |
| Tasks assigned to them in shared workspaces | Kept, unassigned |
| Pull requests they authored on shared tasks | Kept, without their GitHub id |

Shared content stays because it belongs to the team too: deleting a teammate's comments or time
entries would rewrite a shared board. Text another person wrote that mentions the erased person
by name is that person's content and is not edited.

Everything runs in one transaction under the kit's advisory lock. A storage failure rolls it
back, so the next delivery finds the account and tries again; deleting a missing object
succeeds.

## An erased subject never gets an account again

Since v2.37.0 the receiver's handled subjects (`directory_deletion_state`, kept at least 60 days)
guard every path that creates or links an account from a Registre subject:

| Path | Answer for a handled subject |
|---|---|
| Sign-in, bearer provisioning (the first MCP call) and the profile sync, all in porte/oidc (`oidc.Deps.Erased`) | Refused with `permission_denied`, nothing written; the callback shows "this account was deleted" |
| Registre space sync (`workspace.Directory.Apply`), which pre-provisions members | The member is skipped: no user, no identity, no membership |

An account re-created anyway, by a call that raced the erasure, is erased by the next push or
poll of the same notice, with no restart: porte asks `pg.Store.HoldsIdentity` whether an identity
of the issuer holds a handled subject again (`directory.Deletions.Holds`). That check matches
the issuer exactly as `OIDC_ISSUER` spells it, which is how sign-in and the space sync store it.

Tests: `TestAHandledSubjectIsRefusedAnAccount` (`main`),
`TestARepushedNoticeErasesAnAccountRecreatedAfterTheErasure` and
`TestTheSpaceSyncSkipsAnErasedMember` (`modules/workspace`).

## The proof

`modules/workspace/erasure_test.go` seeds a person with a row in every table above, erases them
through `EraseAccount`, then scans every text and JSON column of every table for their email,
name, Registre subject and GitHub id, and asserts there is none (`antenne_outbox` included). It queues unsent events through
`outbox.Enqueue` before the erasure, so a queued event naming the person fails it too. It
also asserts the shared rows survive with an empty author and another person's queued event
stays queued. A new table that stores one of these values fails the test.

Back to the [documentation index](README.md).
