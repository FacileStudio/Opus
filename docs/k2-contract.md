# Phase 1 K2 — frontend auth re-plumb: locked contract

Cold-start contract for the K2 parallel work. The keystone files are locked by the
coordinator and MUST NOT be edited by workers:

- `apps/web/src/lib/auth-client.ts` — the type surface everything imports
- `apps/web/src/lib/permissions.ts`
- `apps/web/src/types/user.ts`, `apps/web/src/types/workspace/index.ts`, `apps/web/src/types/workspace-user/index.ts`
- `packages/libs/src/hono.ts` (CSRF header on the fetch wrapper)

Workers compile against these. Endpoint paths below are binding for the Go worker
and for the web workers that call them.

## Transport rules (porte)

- Session cookie is HttpOnly (`opus_session`); browser sends it automatically.
- **Cookie-authenticated MUTATION** (POST/PATCH/DELETE) MUST carry `X-Facile-CSRF`
  (any value). Bearer-authenticated requests are exempt.
- Bearer token (from `{token, user}`) is sent as `Authorization: Bearer <token>`
  and kept for the fetch wrapper.

## Auth endpoints (Go, all exist or NEW-marked)

| Method | Path | Body | Response |
|---|---|---|---|
| GET | `/api/auth/me` | — | `{token, user}` |
| POST | `/api/auth/login` | `{email, password}` | `{token, user}` |
| POST | `/api/auth/register` | `{email, name, password}` | 201 `{token, user}` |
| POST | `/api/auth/logout` | — | `{loggedOut: true}` |
| GET | `/api/auth/oidc` | — | 302 → IdP |
| GET | `/api/config` | — | legacy config (proxied; unchanged) |
| **PATCH** | `/api/auth/me` | `{name?, color?, locale?}` | `{token, user}` |
| **POST** | `/api/auth/change-password` | `{currentPassword, newPassword}` | `{message}` |

`user` shape (int64 id on the wire): `{id, name, email, emailVerified, image, color, locale, createdAt, updatedAt}`.

## Workspace endpoints (Go)

All int64 ids on the wire. `Member` = `{id, name, email, image, role}` with
`role ∈ owner|admin|member`. `Workspace` = `{id, name, slug, logo, description, createdAt}`.

| Method | Path | Body | Response |
|---|---|---|---|
| GET | `/api/workspace` | — | `Workspace[]` (member of) |
| **POST** | `/api/workspace` | `{name, slug?, description?}` | 201 `Workspace` (creator becomes `owner` member) |
| **PATCH** | `/api/workspace/{id}` | `{name?, slug?, description?}` | `Workspace` |
| **DELETE** | `/api/workspace/{id}` | — | `{message}` |
| GET | `/api/workspace/{id}/members` | — | `Member[]` |
| **PATCH** | `/api/workspace/{id}/members/{userId}` | `{role}` | `Member` |
| **DELETE** | `/api/workspace/{id}/members/{userId}` | — | `{message}` |
| **GET** | `/api/workspace/{id}/full` | — | `{workspace, members, invitations}` |
| **GET** | `/api/workspace/{id}/members/me` | — | `Member` (current user's row) |
| **POST** | `/api/workspace/{id}/members` | `{email, role?}` | 201 `Invitation` |
| **GET** | `/api/workspace/{id}/invitations` | — | `Invitation[]` |
| **GET** | `/api/invitation` | — | `Invitation[]` (user's own) |
| **GET** | `/api/invitation/{id}` | — | `InvitationDetail` |
| **POST** | `/api/invitation/{id}/accept` | — | `{workspace, message}` |
| **POST** | `/api/invitation/{id}/reject` | — | `{message}` |
| **DELETE** | `/api/invitation/{id}` | — | `{message}` (cancel) |

`Invitation` = `{id, email, role, status, createdAt, expiresAt, workspaceId, workspaceName, inviterName, inviter Image?}`.
`InvitationDetail` = `{id, email, role, status, workspace: Workspace, inviter: {name, image}}` (the public accept page uses this — it already works via the legacy `client.invitation.public/:id` proxy; keep it).
`status ∈ pending|accepted|rejected|canceled`. `expiresAt` = hourly-redreshable; default 7 days from created (better-auth default).

## Member shape normalization (keystone + Go, binding)

The existing components read a **nested** member shape across ~20 files:
`member.userId` (int64), `member.role`, and `member.user.{name,image,email,id}`. The Go
workspace module returns a flat `Member {id,name,email,image,role}`. The keystone MUST
expose `listMembers`/`getFullOrganization`/`getActiveMember` members as
`{userId, role, user: {id, name, email, image}}` so those call sites compile unchanged.
`getFullOrganization` returns `{organization, members, invitations}`; `listMembers` returns
`{members, totalCount?}`.

## Frontend surface (auth-client.ts — locked)

- `useSession()` → `{data: {user}|undefined, isPending}` (module store)
- `getSession()` → `{data: {user}|null}`
- `signIn.email({email,password})`, `signUp.email({email,password,name})`, `signOut()`
- `updateUser({name?,color?,locale?})`, `changePassword({currentPassword,newPassword})`
- `organization.*` (below) — reads/writes the Go workspace endpoints
- `getLastUsedLoginMethod()` → `"oidc" | "password" | null` (best-effort from config)
- CUT (never re-added): `signIn.anonymous`, `signIn.social`, `signIn.oauth2`, `signIn.emailOtp`,
  `emailOtp`, `apiKey.*`, `device.*`, `magicLink`, `organization.hasPermission` (server),
  `organization.checkRolePermission` (server). `hasPermission`/`checkRolePermission` return
  client-side booleans from the `Member.role` only (see permissions.ts).

`organization.*`:
- `list()`, `create({name, slug?, description?, logo?})`, `update({organizationId, data})`,
  `delete({organizationId})`
- `listMembers({query: {organizationId, limit?, offset?}})`
- `getFullOrganization({query: {organizationId, membersLimit?}})` → `{workspace, members}`
- `getActiveMember()`, `updateMemberRole({memberId, organizationId, role})`, `removeMember({organizationId, memberIdOrEmail})`
- `inviteMember({organizationId, email, role?})`
- `listInvitations(...)`, `listUserInvitations()`, `acceptInvitation({invitationId})`,
  `rejectInvitation({invitationId})`, `cancelInvitation({invitationId})`
- `setActive({organizationId})` → client-side only (localStorage), no round-trip

## Web worker file ownership

**Worker web-auth** (routes, provider, auth forms, cuts):
- `apps/web/src/components/providers/auth-provider/**`
- `apps/web/src/routes/auth/**, apps/web/src/routes/auth.tsx`
- `apps/web/src/routes/_layout/_authenticated.tsx`, `apps/web/src/routes/_layout/_authenticated/dashboard/index.tsx`
- `apps/web/src/components/auth/**`
- `apps/web/src/components/onboarding/onboarding-flow.tsx`
- `apps/web/src/hooks/use-locale.ts`, `apps/web/src/hooks/use-project-websocket.ts`
- `apps/web/src/hooks/mutations/{use-sign-in,use-sign-up,use-sign-out,use-change-password,use-send-password-reset,use-update-user-profile}.ts`
- DELETE: `apps/web/src/hooks/queries/use-get-api-keys.ts`, `apps/web/src/hooks/mutations/api-key/**`,
  `apps/web/src/routes/device/**`, `apps/web/src/fetchers/oauth/**`, `apps/web/src/fetchers/api-key/**` (if any),
  `apps/web/src/fetchers/{slack,discord,telegram,generic-webhook,github,gitea}-integration/**` fetchers + their hooks
  and routes are Phase-3+ (leave them; they do not import authClient).

**Worker web-workspace** (fetchers, workspace/team hooks, modals):
- `apps/web/src/fetchers/workspace/**, fetchers/workspace-user/**`
- `apps/web/src/hooks/queries/workspace/**, hooks/queries/workspace-users/**, hooks/queries/invitation/**`
- `apps/web/src/hooks/mutations/workspace/**, hooks/mutations/workspace-user/**`
- `apps/web/src/hooks/use-workspace-permission.ts`
- `apps/web/src/components/space-switcher.tsx`, `components/team/**`, `components/shared/modals/create-space-modal.tsx`
- `apps/web/src/routes/_layout/_authenticated/dashboard/invitations.tsx`,
  `apps/web/src/routes/_layout/_authenticated/invitations.tsx`,
  `apps/web/src/routes/_layout/_authenticated/dashboard/workspace/create.tsx`,
  `apps/web/src/routes/invitation/accept.$inviteId.tsx`

**Worker go**:
- `apps/api/**` — new endpoints, invitation schema, tests, legacy proxy auth bridge
- `apps/api-legacy/src/utils/authenticate-api-request.ts` (+ nothing else in api-legacy)

## Gate

`go build ./... && go vet ./... && go test ./...` · `filet check .` from `apps/api` ·
`pnpm --filter @opus/web typecheck` · `pnpm --filter @opus/web build`.