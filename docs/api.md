# Opus: API

The HTTP surface of the Go API, grouped by module. Every path below is relative to `/api` unless
it starts with another root.

The spec the binary serves is the authority. Each module declares its routes in its
`documentation.go`, and `router_test.go` fails when a module is missing:

- `GET /docs`: the interactive reference
- `GET /docs/openapi.json`: the OpenAPI document

## Credentials

Authentication is porte's. A request carries one of:

| Credential | How to send it |
|---|---|
| Session cookie | `__Host-session`, `HttpOnly`, set by the single sign-on callback |
| Session token | `Authorization: Bearer <token>`. A personal API token is a session with a label, so it works everywhere a login does |
| Registre machine JWT | `Authorization: Bearer <jwt>`, when `OIDC_MACHINE_AUDIENCE` is set. It needs `opus.read` for GET, HEAD and OPTIONS and `opus.write` for the rest |

A cookie-authenticated request that is not GET, HEAD or OPTIONS must also send `X-Facile-CSRF`
with any non-empty value; bearer requests need no header. Routes scoped to a workspace check
membership, so a valid credential for the wrong workspace gets 403.

Errors are JSON: `{"error": {"code": "...", "message": "..."}}`. Request bodies are decoded
strictly, so an unknown field answers 400. IDs in request bodies may be JSON numbers or decimal
strings (`internal/wireid`). Send `X-Opus-Window-Id` from a browser tab so realtime broadcasts
skip the tab that made the change.

## Root routes

| Method | Path | What it does |
|---|---|---|
| GET, HEAD | `/health`, `/api/health` | Liveness |
| GET | `/ready`, `/api/ready` | Readiness: pings the database, 503 until it answers |
| GET | `/version` | The version baked into the binary |
| GET | `/docs`, `/docs/openapi.json` | The API reference |

## Auth

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/auth/config` | public | The sign-in methods on offer |
| GET | `/auth/oidc` | public | Start single sign-on with Registre |
| GET | `/auth/oidc/callback` | public | Finish it, set the cookie, redirect to `OIDC_SUCCESS_URL` |
| POST | `/auth/oidc/exchange` | public | Trade a CLI device-grant token for a session. Only with `OIDC_CLI_AUDIENCE` |
| POST | `/auth/logout` | session | End the current session |
| POST | `/auth/backchannel-logout` | logout token | Registre's back-channel logout: a signed `logout_token` revokes every session the subject holds |
| GET | `/auth/me` | session | The caller's profile |
| PATCH | `/auth/me` | session | Update the caller's own profile |
| GET | `/auth/api-tokens` | session | List personal API tokens |
| POST | `/auth/api-tokens` | session | Mint one |
| DELETE | `/auth/api-tokens/{id}` | session | Revoke one |
| GET | `/config` | public | Public settings for the client, including `registreSpacesUrl`, `customOAuthLogoutUrl` (Registre's `/logout`) and `githubAppInstallUrl` (where a GitHub account installs the Facile Opus App; empty when the App is off) |

## Registre

| Method | Path | Auth | What it does |
|---|---|---|---|
| POST | `/registre/sync` | public | Nudge a space sync. Only with OIDC |
| POST | `/registre/deletions` | Registre signature | Receive a deletion notice. Only with OIDC. See [erasure.md](erasure.md) |

## Workspaces and invitations

| Method | Path | What it does |
|---|---|---|
| GET | `/workspace` | The caller's workspaces, personal first |
| POST | `/workspace` | Create one. 409 once the Registre import has succeeded |
| PATCH | `/workspace/{workspaceId}` | Update one. 409 on a linked workspace |
| DELETE | `/workspace/{workspaceId}` | Delete one. 409 on a linked or personal workspace |
| GET | `/workspace/{workspaceId}/full` | A workspace and its members |
| GET | `/workspace/{workspaceId}/members` | Members |
| GET | `/workspace/{workspaceId}/members/me` | The caller's membership |
| PATCH | `/workspace/{workspaceId}/members/{userId}` | Change a role |
| DELETE | `/workspace/{workspaceId}/members/{memberIdOrEmail}` | Remove a member |
| POST | `/workspace/{workspaceId}/members` | Invite by email |
| GET | `/workspace/{workspaceId}/invitations` | A workspace's invitations |
| GET | `/workspace/{workspaceId}/logo` | The logo, public |
| POST, DELETE | `/workspace/{workspaceId}/logo` | Upload or remove the logo |
| GET | `/invitation` | The caller's pending invitations |
| GET | `/invitation/{invitationId}` | One invitation |
| POST | `/invitation/{invitationId}/accept` | Accept |
| POST | `/invitation/{invitationId}/reject` | Reject |
| DELETE | `/invitation/{invitationId}` | Cancel |

## Projects, columns and workflow rules

| Method | Path | What it does |
|---|---|---|
| GET | `/project?workspaceId=` | Projects in a workspace |
| POST | `/project` | Create one |
| GET, PUT, DELETE | `/project/{id}` | Read, update, delete one |
| PUT | `/project/{id}/archive`, `/project/{id}/unarchive` | Archive, restore |
| GET | `/public-project/{id}` | A project marked public, no auth |
| GET, POST | `/column/{projectId}` | List, create columns |
| PUT, DELETE | `/column/{id}` | Update, delete one |
| PUT | `/column/reorder/{projectId}` | Reorder |
| GET, PUT | `/workflow-rule/{projectId}` | List, upsert rules |
| DELETE | `/workflow-rule/{id}` | Delete one |

## Tasks

| Method | Path | What it does |
|---|---|---|
| GET | `/task/tasks/{projectId}` | All tasks of a project |
| GET | `/task/me` | Tasks assigned to the caller |
| GET | `/task/{id}` | One task |
| POST | `/task/{projectId}` | Create a task |
| PUT | `/task/{id}` | Update every field |
| DELETE | `/task/{id}` | Delete a task |
| PATCH | `/task/bulk` | Bulk update |
| PUT | `/task/move/{id}` | Move to another project |
| PUT | `/task/status/{id}`, `/task/priority/{id}`, `/task/actor/{id}`, `/task/due-date/{id}`, `/task/title/{id}`, `/task/description/{id}` | Update one field |
| GET | `/task-relation/{taskId}` | Relations of a task |
| POST | `/task-relation` | Relate two tasks |
| DELETE | `/task-relation/{id}` | Remove a relation |
| GET | `/external-link/task/{taskId}` | Always an empty list, kept for the client |

The single-field routes let a board drag or an inline edit send the smallest write and record a
precise activity entry.

## Labels, comments and activity

| Method | Path | What it does |
|---|---|---|
| GET | `/label/workspace/{workspaceId}` | Labels of a workspace |
| GET | `/label/task/{taskId}` | Labels on a task |
| POST | `/label` | Create one |
| GET, PUT, DELETE | `/label/{id}` | Read, update, delete one |
| PUT, DELETE | `/label/{id}/task` | Attach to or detach from a task |
| GET, POST | `/comment/{taskId}` | List, add comments |
| PUT, DELETE | `/comment/{id}` | Edit, delete one |
| GET | `/activity/{taskId}` | Activity feed of a task |
| POST | `/activity/create` | Record an activity |
| POST, PUT, DELETE | `/activity/comment` | Add, edit, delete a comment in the feed. The web client comments here |

## Time entries, share links, search

| Method | Path | What it does |
|---|---|---|
| GET | `/time-entry/task/{taskId}` | Entries of a task |
| POST | `/time-entry` | Create one |
| GET, PUT, DELETE | `/time-entry/{id}` | Read, update, delete one. Entries mirrored from Sablier answer 409 to edits |
| GET, POST | `/share-link` | List, create share links of a project |
| PUT, DELETE | `/share-link/{id}` | Update, delete one |
| GET | `/shared-project/{token}` | A project behind a share link, no auth |
| GET | `/search` | Search tasks, projects, workspaces and comments the caller can see |

## Notifications

| Method | Path | What it does |
|---|---|---|
| GET, POST | `/notification` | List, create |
| PATCH | `/notification/{id}/read`, `/notification/{id}/unread` | Mark one |
| DELETE | `/notification/{id}` | Delete one |
| PATCH | `/notification/read-all` | Mark all read |
| DELETE | `/notification/clear-all` | Clear all |
| GET, PUT | `/notification-preferences` | Global preferences |
| PUT, DELETE | `/notification-preferences/workspaces/{workspaceId}` | Per-workspace rule |

## Assets

| Method | Path | What it does |
|---|---|---|
| PUT | `/task/image-upload/{id}` | Start a task image upload. Returns the upload URL and key |
| POST | `/task/image-upload/{id}/finalize` | Register the uploaded object as an asset |
| PUT | `/asset/upload/*` | Local storage only: receives the bytes when S3 is not configured. Takes no session: the signed URL from `/task/image-upload/{id}` is the authorization (15 minutes, size cap, never overwrites). 401 unsigned, 403 bad or expired signature, 413 over the cap, 409 if the object exists |
| GET | `/asset/{id}` | Download. Workspace members, or `?share_token=` when the project is not public |

## GitHub

| Method | Path | What it does |
|---|---|---|
| POST | `/integrations/github/webhook` | GitHub webhook, verified with `GITHUB_WEBHOOK_SECRET` |
| GET | `/project/{projectId}/task/{taskId}/pull-requests` | Pull requests linked to a task |
| GET | `/project/{projectId}/github-repos` | Repositories linked to the project. Workspace members |
| POST | `/project/{projectId}/github-repos` | Link `{"repo": "owner/repo"}`. Workspace owner or admin, a GitHub account linked in Registre, and write or admin on the repository. 409 when already linked |
| DELETE | `/project/{projectId}/github-repos/{repoId}` | Unlink. Workspace owner or admin. Works without the App |
| GET | `/project/{projectId}/github-issues?state=open\|closed` | Issues of every linked repository, pull requests excluded, at most 100 per repository, cached 60 s. Workspace members. A repository GitHub refuses is listed under `unavailable` with the reason |

The linked-repository and issues routes answer 412 with a plain reason when the Facile Opus App
is not configured, when the caller has no GitHub account on file (POST), or when the App is not
installed on the repository (POST).

## Sync and Ideas

| Method | Path | What it does |
|---|---|---|
| POST | `/sync/reconcile` | Re-send every syncable project and task to Antenne. Owner or admin of a workspace; answers 202 |
| GET | `/ideas` | Bulle's ideas for the active workspace. Only with `BULLE_URL` and OIDC |
| POST | `/ideas/generate`, `/ideas/{id}/run`, `/ideas/{id}/dismiss` | Generate, run, dismiss an idea |
| GET | `/ideas/runs/{run}` | A run's status |
| POST | `/ideas/runs/{run}/approvals/{call}` | Approve or refuse a step |

## MCP

The MCP server lives at the root. It is on only when `OIDC_ISSUER` and `MCP_RESOURCE` are set.

| Method | Path | Auth | What it does |
|---|---|---|---|
| POST | `/mcp` | Registre access token bound to `MCP_RESOURCE` | Model Context Protocol, Streamable HTTP, stateless |
| GET | `/.well-known/oauth-protected-resource/mcp` | none | RFC 9728 metadata: the resource, Registre as the authorization server, `opus.read` and `opus.write` |
| GET | `/api/mcp/pending` | session | Deletes an MCP client asked for and could not confirm with you |
| POST | `/api/mcp/pending/{id}/confirm` | session | Run one, once. 412 when the target changed since the request |
| POST | `/api/mcp/pending/{id}/dismiss` | session | Close one without running it |

`GET /mcp` answers 405. A request without a valid token gets 401 with a `WWW-Authenticate`
header naming the metadata; a write tool called with only `opus.read` gets 403
`insufficient_scope`. A foreign browser `Origin` gets 403. Delegated tokens cannot confirm pending
actions. The tools are listed in the [README](../README.md#mcp).

## WebSocket

| Path | Auth | What it does |
|---|---|---|
| `/ws/{projectId}`, `/api/ws/{projectId}` | session cookie, bearer token or `?token=` | Live updates for one project. `?windowId=` skips echoes to the originating tab |

The caller must be a member of the project's workspace. Broadcasts go through Redis when
`REDIS_URL` is set and stay in memory otherwise.
