# Opus — API

Every HTTP route the Hono app mounts, grouped by module. Generated from
`apps/api/src/index.ts` and the per-feature routers.

Opus serves its own machine-readable spec, and that spec is the authority:

- `GET /api/openapi` — merged OpenAPI 3.0.3 document (feature routes plus the `better-auth`
  surface)
- `GET /api/docs` — interactive Scalar reference

Every path below is relative to `/api`.

## Authentication

Three credential types are accepted, resolved by `authenticateApiRequest`:

| Credential | How to send it |
|---|---|
| Session cookie | Set by `better-auth` after a browser sign-in |
| Bearer token | `Authorization: Bearer <session-token>` |
| API key | `x-api-key: <key>`, or `Authorization: Bearer <key>` on non-auth routes |

Everything is authenticated except the public routes listed below. Routes scoped to a workspace
additionally run a membership check, so a valid credential for the wrong workspace gets a 403.
API keys are rate-limited to 100 requests per minute.

Errors are JSON `HTTPException` bodies. Send `X-Opus-Window-Id` from a browser client so the
WebSocket fan-out can skip echoing your own changes back to you.

## Public routes

| Method | Path | What it does |
|---|---|---|
| GET | `/health` | Liveness probe, returns `{"status":"ok"}` |
| GET | `/config` | Public feature flags — `ssoOnly`, `disableRegistration`, `disablePasswordRegistration`, `isDemoMode`, `hasSmtp`, `hasGithubSignIn`, `hasGoogleSignIn`, `hasDiscordSignIn`, `hasCustomOAuth`, `hasGuestAccess`, `customOAuthLogoutUrl` |
| GET | `/public-project/:id` | A project marked public |
| GET | `/shared-project/:token` | A project behind a share link token |
| GET | `/invitation/public/:id` | Invitation details before signing in |
| GET | `/asset/:id` | Download an asset. Authorized by workspace membership, or by `?share_token=` when the project is not public |
| GET | `/workspace/:workspaceId/logo` | Workspace logo image, cached for a year |
| GET | `/openapi` | Merged OpenAPI document |
| GET | `/docs` | Scalar API reference |
| POST | `/github-integration/webhook` | GitHub App webhook receiver, verified with `GITHUB_WEBHOOK_SECRET` |
| POST | `/gitea-integration/webhook/:integrationId` | Gitea webhook receiver |

## Authentication routes

`/auth/*` is handled entirely by `better-auth` (sign-in, sign-up, magic link, email OTP, social
callbacks, organization management, API keys, device authorization). Its operations are merged
into `/api/openapi`; read that rather than a list here, since the surface changes with the plugin
set.

Two routes are wrapped by hand:

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/auth/get-session` | optional | Current session, or `null` |
| GET | `/auth/device` | none | Redirects real browser navigations to `<OPUS_CLIENT_URL>/device`, returns the Better Auth JSON payload otherwise. `?ui=1` forces the redirect |

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/oauth/id-token` | session | The `id_token` of the current user's custom OAuth account |

## Workspaces

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/workspace` | session | Workspaces the caller belongs to |
| GET | `/workspace/:workspaceId/members` | member | Members of a workspace |
| POST | `/workspace/:workspaceId/logo` | member | Upload a workspace logo |
| DELETE | `/workspace/:workspaceId/logo` | member | Remove the logo |
| GET | `/invitation/pending` | session | Pending invitations addressed to the caller's email |
| GET | `/invitation/:id` | session | Details of one invitation |

## Projects

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/project?workspaceId=&includeArchived=` | member | Projects in a workspace, with statistics |
| POST | `/project` | member | Create a project. Body: `name`, `workspaceId`, `icon`, `slug` |
| GET | `/project/:id` | member | One project |
| PUT | `/project/:id` | member | Update a project |
| DELETE | `/project/:id` | member | Delete a project |
| PUT | `/project/:id/archive` | member | Archive |
| PUT | `/project/:id/unarchive` | member | Unarchive |

## Columns and workflow rules

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/column/:projectId` | member | Columns of a project |
| POST | `/column/:projectId` | member | Create a column |
| PUT | `/column/reorder/:projectId` | member | Reorder columns |
| PUT | `/column/:id` | member | Update a column |
| DELETE | `/column/:id` | member | Delete a column |
| GET | `/workflow-rule/:projectId` | member | Workflow rules of a project |
| PUT | `/workflow-rule/:projectId` | member | Create or update a rule |
| DELETE | `/workflow-rule/:id` | member | Delete a rule |

## Tasks

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/task/tasks/:projectId` | member | All tasks in a project |
| GET | `/task/:id` | member | One task |
| POST | `/task/:projectId` | member | Create a task |
| PUT | `/task/:id` | member | Update every field |
| PATCH | `/task/bulk` | member | Bulk operations across several tasks |
| DELETE | `/task/:id` | member | Delete a task |
| PUT | `/task/move/:id` | member | Move a task to another project |
| PUT | `/task/status/:id` | member | Update status only |
| PUT | `/task/priority/:id` | member | Update priority only |
| PUT | `/task/actor/:id` | member | Assign or unassign |
| PUT | `/task/due-date/:id` | member | Update due date only |
| PUT | `/task/title/:id` | member | Update title only |
| PUT | `/task/description/:id` | member | Update description only |
| GET | `/task/export/:projectId` | member | Export every task of a project |
| POST | `/task/import/:projectId` | member | Import tasks into a project |
| PUT | `/task/image-upload/:id` | member | Create a presigned S3 upload URL for a description or comment image |
| POST | `/task/image-upload/:id/finalize` | member | Register the uploaded object as a private asset |

The narrow single-field routes exist so a kanban drag or an inline edit sends the smallest possible
write and produces a precise activity entry.

## Comments, activity and relations

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/comment/:taskId` | member | Comments on a task |
| POST | `/comment/:taskId` | member | Add a comment |
| PUT | `/comment/:id` | author | Edit a comment |
| DELETE | `/comment/:id` | author | Delete a comment |
| GET | `/activity/:taskId` | member | Activity feed of a task |
| POST | `/activity/create` | member | Record a system-generated activity |
| POST | `/activity/comment` | member | Comment through the activity feed |
| PUT | `/activity/comment` | author | Edit that comment |
| DELETE | `/activity/comment` | author | Delete that comment |
| GET | `/task-relation/:taskId` | member | Relations of a task |
| POST | `/task-relation` | member | Relate two tasks |
| DELETE | `/task-relation/:id` | member | Remove a relation |
| GET | `/external-link/task/:taskId` | member | External links attached to a task |

## Labels

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/label/workspace/:workspaceId` | member | Labels in a workspace |
| GET | `/label/task/:taskId` | member | Labels on a task |
| GET | `/label/:id` | member | One label |
| POST | `/label` | member | Create a label |
| PUT | `/label/:id` | member | Update a label |
| DELETE | `/label/:id` | member | Delete a label |
| PUT | `/label/:id/task` | member | Attach a label to a task |
| DELETE | `/label/:id/task` | member | Detach it |

## Time entries

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/time-entry/task/:taskId` | member | Time entries on a task |
| GET | `/time-entry/:id` | member | One entry |
| POST | `/time-entry` | member | Create an entry |
| PUT | `/time-entry/:id` | member | Update an entry |

## Search

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/search` | session | Search across tasks, projects, workspaces, comments and activities |

## Notifications

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/notification` | session | Notifications for the caller |
| POST | `/notification` | session | Create a notification for a user |
| PATCH | `/notification/:id/read` | session | Mark one as read |
| PATCH | `/notification/read-all` | session | Mark all as read |
| DELETE | `/notification/clear-all` | session | Clear all |
| GET | `/notification-preferences` | session | Global delivery preferences |
| PUT | `/notification-preferences` | session | Update global preferences |
| PUT | `/notification-preferences/workspaces/:workspaceId` | session | Create or update a per-workspace rule |
| DELETE | `/notification-preferences/workspaces/:workspaceId` | session | Delete that rule |

Provider secrets submitted here are encrypted at rest with `NOTIFICATION_SECRET_ENCRYPTION_KEY`.

## Share links

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/share-link` | member | Share links of a project |
| POST | `/share-link` | member | Create one |
| PUT | `/share-link/:id` | member | Update one |
| DELETE | `/share-link/:id` | member | Delete one |

Tokens are consumed by the public `GET /shared-project/:token` and by `?share_token=` on
`GET /asset/:id`.

## Integrations

Every per-project integration follows the same four-route shape.

| Method | Path | What it does |
|---|---|---|
| GET | `/<integration>/project/:projectId` | Read the integration config |
| POST | `/<integration>/project/:projectId` | Create or replace it |
| PATCH | `/<integration>/project/:projectId` | Update settings |
| DELETE | `/<integration>/project/:projectId` | Remove it |

`<integration>` is one of `github-integration`, `gitea-integration`,
`generic-webhook-integration`, `discord-integration`, `slack-integration`,
`telegram-integration`.

The two source-forge integrations add a few routes:

| Method | Path | What it does |
|---|---|---|
| GET | `/github-integration/app-info` | GitHub App configuration as the server sees it |
| GET | `/github-integration/repositories` | Repositories the installation can reach |
| POST | `/github-integration/verify` | Verify the App installation on a repository |
| POST | `/github-integration/import-issues` | Import GitHub issues as tasks |
| POST | `/gitea-integration/repositories` | Repositories reachable with a supplied token |
| POST | `/gitea-integration/verify` | Verify a token and repository |
| POST | `/gitea-integration/import-issues` | Import Gitea issues as tasks |

## Workspace webhooks

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/workspace-webhook/:workspaceId` | member | List webhooks |
| POST | `/workspace-webhook/:workspaceId` | member | Create a webhook |
| PATCH | `/workspace-webhook/:workspaceId/:webhookId` | member | Update it |
| DELETE | `/workspace-webhook/:workspaceId/:webhookId` | member | Delete it |
| POST | `/workspace-webhook/:workspaceId/:webhookId/test` | member | Send a test ping |

## Nook pool

| Method | Path | Auth | What it does |
|---|---|---|---|
| GET | `/nook-pool/:workspaceId` | member | Current pool configuration |
| POST | `/nook-pool/:workspaceId` | member | Save instance URL, secret and enabled flag |
| DELETE | `/nook-pool/:workspaceId` | member | Delete the configuration and disconnect |
| POST | `/nook-pool/:workspaceId/test` | member | Test the connection to a pool instance |
| POST | `/nook-pool/:workspaceId/sync` | member | Push every existing project and task to the pool |
| GET | `/nook-pool/:workspaceId/events` | member | Per-event toggles |
| PUT | `/nook-pool/:workspaceId/events` | member | Update those toggles |

Emitted and accepted events: `project.created`, `project.updated`, `project.deleted`,
`task.created`, `task.updated`, `task.deleted`. See
[architecture.md](architecture.md) for the outbox and idempotency machinery.

## MCP

| Method | Path | Auth | What it does |
|---|---|---|---|
| POST | `/mcp/register` | none | Dynamic client registration |
| GET | `/mcp/authorize` | none | Authorization endpoint |
| POST | `/mcp/authorize/callback` | none | Authorization callback |
| POST | `/mcp/token` | none | Token endpoint |
| GET | `/.well-known/oauth-protected-resource/api/mcp` | none | Protected resource metadata |
| GET | `/.well-known/oauth-authorization-server/api` | none | Authorization server metadata |

The two `.well-known` documents are also served by nginx at the root, populated at container start
by `apps/web/env.sh` from `OPUS_API_URL`.

## WebSocket

| Path | Auth | What it does |
|---|---|---|
| `/ws/:projectId` | session, bearer or API key | Live project updates. Optional `?windowId=` deduplicates echoes back to the originating tab |

The caller must be a member of the project's workspace. Broadcasts go through Redis when it is
configured and in-memory otherwise.
