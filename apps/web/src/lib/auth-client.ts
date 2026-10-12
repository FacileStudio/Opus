import { useSyncExternalStore } from "react";
import { getApiUrl } from "@/fetchers/get-api-url";
import { roleHas } from "./permissions";

/**
 * Thin fetch client against the Go backend. This is the single auth + workspace
 * surface the web app imports — every other module reads its types from here,
 * so the shapes below are the contract. IDs arrive as int64 on the wire and
 * are normalized to strings here: the frontend treats them as opaque.
 *
 * Transport follows porte: the session cookie is HttpOnly (sent automatically
 * with credentials: include), and every cookie-authenticated mutation carries
 * the X-Facile-CSRF header (any value). The bearer token from {token, user} is
 * kept and sent when present; bearer requests are exempt from the CSRF rule.
 */

export type User = {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string;
  color: string;
  locale: string;
  createdAt: string;
  updatedAt: string;
};

export type Session = {
  user: User;
};

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  logo: string | null;
  description: string | null;
  createdAt: string;
};

export type WorkspaceMember = {
  userId: string;
  role: "owner" | "admin" | "member";
  user: {
    id: string;
    name: string;
    email: string;
    image: string | null;
  };
};

export type WorkspaceInvitation = {
  id: string;
  email: string;
  role: "owner" | "admin" | "member";
  status: "pending" | "accepted" | "rejected" | "canceled";
  createdAt: string;
  expiresAt: string | null;
  workspaceId: string;
  workspaceName: string;
  inviterName: string;
};

export type ApiError = {
  message: string;
  status?: number;
};

type ApiResult<T> = Promise<{ data: T | null; error: ApiError | null }>;

function apiUrl(path: string): string {
  return getApiUrl(path);
}

const CSRF_HEADER = "X-Facile-CSRF";

// Session store: one source of truth for the authenticated user, hydrated by
// AuthProvider from GET /auth/me. useSyncExternalStore keeps every consumer on
// the same snapshot.
let sessionState: { user: User | null; isPending: boolean } = {
  user: null,
  isPending: true,
};
const sessionListeners = new Set<() => void>();

function setSessionState(next: Partial<typeof sessionState>) {
  sessionState = { ...sessionState, ...next };
  for (const listener of sessionListeners) {
    listener();
  }
}

function subscribeSession(listener: () => void) {
  sessionListeners.add(listener);
  return () => sessionListeners.delete(listener);
}

async function request<T>(
  path: string,
  init: RequestInit = {},
  options: { mutation?: boolean } = {},
): Promise<{ data: T; error: null } | { data: null; error: ApiError }> {
  const headers = new Headers(init.headers);
  if (init.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  if (options.mutation) {
    headers.set(CSRF_HEADER, "1");
  }
  // Bearer token kept for the fetch wrapper.
  const token = localStorage.getItem("opus.access_token");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const response = await fetch(apiUrl(path), {
    ...init,
    headers,
    credentials: "include",
  });
  // porte answers 401 when there is no valid credential and 403 when there is
  // one but it lacks permission. A 401 therefore means the session is gone:
  // clear the store so the UI stops holding a ghost user instead of waiting
  // for a reload. The next getSession on focus confirms or restores it.
  if (
    response.status === 401 &&
    path !== "/auth/login" &&
    path !== "/auth/register"
  ) {
    setSessionState({ user: null, isPending: false });
  }
  if (!response.ok) {
    let message = response.statusText || "Request failed";
    try {
      const body = await response.json();
      message = body?.message || body?.error || message;
    } catch {
      // non-JSON error body
    }
    return { data: null, error: { message, status: response.status } };
  }
  if (response.status === 204) {
    return { data: undefined as T, error: null };
  }
  const data = (await response.json()) as T;
  return { data, error: null };
}

function toStr(value: number | string | undefined | null): string {
  return value === undefined || value === null ? "" : String(value);
}

// hasRolePermission is the client-side role gate the UI uses to render
// controls. The server owns enforcement; this only answers what the member
// role the browser already holds allows. The role→capability map lives in
// permissions.ts (imported below), the single source the better-auth access
// plugin used to be.
function hasRolePermission(
  role: WorkspaceMember["role"],
  permissions: Record<string, string[]> = {},
): boolean {
  return Object.entries(permissions).every(([resource, actions]) =>
    actions.every((action) => roleHas(role, resource, action)),
  );
}

function mapMember(member: {
  id: number | string;
  name: string;
  email: string;
  image?: string | null;
  role: string;
}): WorkspaceMember {
  return {
    userId: toStr(member.id),
    role: member.role as WorkspaceMember["role"],
    user: {
      id: toStr(member.id),
      name: member.name,
      email: member.email,
      image: member.image ?? null,
    },
  };
}

function mapWorkspace(workspace: {
  id: number | string;
  name: string;
  slug: string;
  logo?: string | null;
  description?: string | null;
  createdAt?: string;
}): Workspace {
  return {
    id: toStr(workspace.id),
    name: workspace.name,
    slug: workspace.slug,
    logo: workspace.logo ?? null,
    description: workspace.description ?? null,
    createdAt: workspace.createdAt ?? "",
  };
}

function mapInvitation(invitation: {
  id: number | string;
  email: string;
  role: string;
  status: string;
  createdAt?: string;
  expiresAt?: string | null;
  workspaceId: number | string;
  workspaceName?: string;
  inviterName?: string;
}): WorkspaceInvitation {
  return {
    id: toStr(invitation.id),
    email: invitation.email,
    role: invitation.role as WorkspaceInvitation["role"],
    status: invitation.status as WorkspaceInvitation["status"],
    createdAt: invitation.createdAt ?? "",
    expiresAt: invitation.expiresAt ?? null,
    workspaceId: toStr(invitation.workspaceId),
    workspaceName: invitation.workspaceName ?? "",
    inviterName: invitation.inviterName ?? "",
  };
}

function mapSessionUser(user: {
  id: number | string;
  name: string;
  email: string;
  emailVerified?: boolean;
  image?: string;
  color?: string;
  locale?: string;
  createdAt?: string;
  updatedAt?: string;
}): User {
  return {
    id: toStr(user.id),
    name: user.name,
    email: user.email,
    emailVerified: user.emailVerified ?? false,
    image: user.image ?? "",
    color: user.color ?? "",
    locale: user.locale ?? "",
    createdAt: user.createdAt ?? "",
    updatedAt: user.updatedAt ?? "",
  };
}

export const authClient = {
  useSession: () => {
    const state = useSyncExternalStore(
      subscribeSession,
      () => sessionState,
      () => ({ user: null, isPending: true }),
    );
    return {
      data: state.user ? { user: state.user } : undefined,
      isPending: state.isPending,
    };
  },

  async getSession(): Promise<{ data: Session | null }> {
    const result = await request<{ token: string; user: User }>("/auth/me");
    if (result.error) {
      return { data: null };
    }
    const user = mapSessionUser(result.data.user);
    return { data: { user } };
  },

  async signInEmail(input: {
    email: string;
    password: string;
  }): Promise<{ data: Session | null; error: ApiError | null }> {
    const result = await request<{ token: string; user: User }>(
      "/auth/login",
      {
        method: "POST",
        body: JSON.stringify(input),
      },
      { mutation: true },
    );
    if (result.error) {
      return { data: null, error: result.error };
    }
    localStorage.setItem("opus.access_token", result.data.token);
    const user = mapSessionUser(result.data.user);
    setSessionState({ user, isPending: false });
    return { data: { user }, error: null };
  },

  async signUpEmail(input: {
    email: string;
    password: string;
    name: string;
  }): Promise<{ data: Session | null; error: ApiError | null }> {
    const result = await request<{ token: string; user: User }>(
      "/auth/register",
      {
        method: "POST",
        body: JSON.stringify(input),
      },
      { mutation: true },
    );
    if (result.error) {
      return { data: null, error: result.error };
    }
    localStorage.setItem("opus.access_token", result.data.token);
    const user = mapSessionUser(result.data.user);
    setSessionState({ user, isPending: false });
    return { data: { user }, error: null };
  },

  async signOut(): Promise<{ data: null; error: ApiError | null }> {
    const result = await request<{ loggedOut: boolean }>(
      "/auth/logout",
      {
        method: "POST",
      },
      { mutation: true },
    );
    localStorage.removeItem("opus.access_token");
    setSessionState({ user: null, isPending: false });
    return { data: null, error: result.error };
  },

  async updateUser(input: {
    name?: string;
    locale?: string;
    color?: string;
  }): Promise<{ data: Session | null; error: ApiError | null }> {
    const result = await request<{ token: string; user: User }>(
      "/auth/me",
      {
        method: "PATCH",
        body: JSON.stringify(input),
      },
      { mutation: true },
    );
    if (result.error) {
      return { data: null, error: result.error };
    }
    const user = mapSessionUser(result.data.user);
    setSessionState({ user, isPending: false });
    return { data: { user }, error: null };
  },

  async changePassword(input: {
    currentPassword: string;
    newPassword: string;
  }): Promise<{ data: null; error: ApiError | null }> {
    const result = await request<{ message: string }>(
      "/auth/change-password",
      {
        method: "POST",
        body: JSON.stringify(input),
      },
      { mutation: true },
    );
    return { data: null, error: result.error };
  },

  getLastUsedLoginMethod: (): "oidc" | "password" | null =>
    (localStorage.getItem("opus.last_login_method") as "oidc" | "password") ||
    null,

  organization: {
    async list(): ApiResult<Workspace[]> {
      const result =
        await request<
          Array<{
            id: number | string;
            name: string;
            slug: string;
            logo?: string | null;
            description?: string | null;
            createdAt?: string;
          }>
        >("/workspace");
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: result.data.map(mapWorkspace), error: null };
    },

    async listMembers(input: {
      query: { organizationId?: string; limit?: number; offset?: number };
    }): ApiResult<{ members: WorkspaceMember[] }> {
      const organizationId = input?.query?.organizationId;
      if (!organizationId) {
        return { data: { members: [] }, error: null };
      }
      const result = await request<
        Array<{
          id: number | string;
          name: string;
          email: string;
          image?: string | null;
          role: string;
        }>
      >(`/workspace/${organizationId}/members`);
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: { members: result.data.map(mapMember) }, error: null };
    },

    async getFullOrganization(input: {
      query: { organizationId?: string; membersLimit?: number };
    }): ApiResult<{
      organization: Workspace | null;
      members: WorkspaceMember[];
      invitations: WorkspaceInvitation[];
    }> {
      const organizationId = input?.query?.organizationId;
      if (!organizationId) {
        return {
          data: { organization: null, members: [], invitations: [] },
          error: null,
        };
      }
      const result = await request<{
        workspace: {
          id: number | string;
          name: string;
          slug: string;
          logo?: string | null;
          description?: string | null;
          createdAt?: string;
        };
        members: Array<{
          id: number | string;
          name: string;
          email: string;
          image?: string | null;
          role: string;
        }>;
        invitations?: Array<{
          id: number | string;
          email: string;
          role: string;
          status: string;
          createdAt?: string;
          expiresAt?: string | null;
          workspaceId: number | string;
          workspaceName?: string;
          inviterName?: string;
        }>;
      }>(`/workspace/${organizationId}/full`);
      if (result.error) {
        return { data: null, error: result.error };
      }
      return {
        data: {
          organization: mapWorkspace(result.data.workspace),
          members: result.data.members.map(mapMember),
          invitations: (result.data.invitations ?? []).map(mapInvitation),
        },
        error: null,
      };
    },

    async create(input: {
      name: string;
      slug?: string;
      logo?: string;
      metadata?: { description?: string };
    }): ApiResult<Workspace> {
      const result = await request<{
        id: number | string;
        name: string;
        slug: string;
        logo?: string | null;
        description?: string | null;
        createdAt?: string;
      }>(
        "/workspace",
        {
          method: "POST",
          body: JSON.stringify({
            name: input.name,
            slug: input.slug,
            description: input.metadata?.description,
          }),
        },
        { mutation: true },
      );
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: mapWorkspace(result.data), error: null };
    },

    async update(input: {
      organizationId: string;
      data: {
        name?: string;
        slug?: string;
        logo?: string;
        metadata?: { description?: string };
      };
    }): ApiResult<Workspace> {
      const result = await request<{
        id: number | string;
        name: string;
        slug: string;
        logo?: string | null;
        description?: string | null;
        createdAt?: string;
      }>(
        `/workspace/${input.organizationId}`,
        {
          method: "PATCH",
          body: JSON.stringify({
            name: input.data.name,
            slug: input.data.slug,
            description: input.data.metadata?.description,
          }),
        },
        { mutation: true },
      );
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: mapWorkspace(result.data), error: null };
    },

    async delete(input: {
      organizationId: string;
    }): Promise<{ data: { message: string } | null; error: ApiError | null }> {
      return request<{ message: string }>(
        `/workspace/${input.organizationId}`,
        {
          method: "DELETE",
        },
        { mutation: true },
      );
    },

    async inviteMember(input: {
      organizationId: string;
      email: string;
      role?: "owner" | "admin" | "member";
    }): ApiResult<WorkspaceInvitation> {
      const result = await request<{
        id: number | string;
        email: string;
        role: string;
        status: string;
        createdAt?: string;
        expiresAt?: string | null;
        workspaceId: number | string;
        workspaceName?: string;
        inviterName?: string;
      }>(
        `/workspace/${input.organizationId}/members`,
        {
          method: "POST",
          body: JSON.stringify({
            email: input.email,
            role: input.role ?? "member",
          }),
        },
        { mutation: true },
      );
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: mapInvitation(result.data), error: null };
    },

    async removeMember(input: {
      organizationId: string;
      memberIdOrEmail: string;
    }): Promise<{ data: { message: string } | null; error: ApiError | null }> {
      return request<{ message: string }>(
        `/workspace/${input.organizationId}/members/${input.memberIdOrEmail}`,
        { method: "DELETE" },
        { mutation: true },
      );
    },

    async updateMemberRole(input: {
      memberId: string;
      organizationId: string;
      role: "owner" | "admin" | "member";
    }): ApiResult<WorkspaceMember> {
      const result = await request<{
        id: number | string;
        name: string;
        email: string;
        image?: string | null;
        role: string;
      }>(
        `/workspace/${input.organizationId}/members/${input.memberId}`,
        {
          method: "PATCH",
          body: JSON.stringify({ role: input.role }),
        },
        { mutation: true },
      );
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: mapMember(result.data), error: null };
    },

    async listInvitations(input: {
      query: { organizationId?: string };
    }): ApiResult<WorkspaceInvitation[]> {
      const organizationId = input?.query?.organizationId;
      if (!organizationId) {
        return { data: [], error: null };
      }
      const result = await request<
        Array<{
          id: number | string;
          email: string;
          role: string;
          status: string;
          createdAt?: string;
          expiresAt?: string | null;
          workspaceId: number | string;
          workspaceName?: string;
          inviterName?: string;
        }>
      >(`/workspace/${organizationId}/invitations`);
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: result.data.map(mapInvitation), error: null };
    },

    async listUserInvitations(): ApiResult<WorkspaceInvitation[]> {
      const result =
        await request<
          Array<{
            id: number | string;
            email: string;
            role: string;
            status: string;
            createdAt?: string;
            expiresAt?: string | null;
            workspaceId: number | string;
            workspaceName?: string;
            inviterName?: string;
          }>
        >("/invitation");
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: result.data.map(mapInvitation), error: null };
    },

    async getInvitation(input: {
      query: { id: string };
    }): ApiResult<WorkspaceInvitation> {
      const result = await request<{
        id: number | string;
        email: string;
        role: string;
        status: string;
        createdAt?: string;
        expiresAt?: string | null;
        workspaceId: number | string;
        workspaceName?: string;
        inviterName?: string;
      }>(`/invitation/${input.query.id}`);
      if (result.error) {
        return { data: null, error: result.error };
      }
      return { data: mapInvitation(result.data), error: null };
    },

    async acceptInvitation(input: {
      invitationId: string;
    }): Promise<{ data: { message: string } | null; error: ApiError | null }> {
      return request<{ message: string }>(
        `/invitation/${input.invitationId}/accept`,
        {
          method: "POST",
        },
        { mutation: true },
      );
    },

    async rejectInvitation(input: {
      invitationId: string;
    }): Promise<{ data: { message: string } | null; error: ApiError | null }> {
      return request<{ message: string }>(
        `/invitation/${input.invitationId}/reject`,
        {
          method: "POST",
        },
        { mutation: true },
      );
    },

    async cancelInvitation(input: {
      invitationId: string;
    }): Promise<{ data: { message: string } | null; error: ApiError | null }> {
      return request<{ message: string }>(
        `/invitation/${input.invitationId}`,
        {
          method: "DELETE",
        },
        { mutation: true },
      );
    },

    async setActive(input: { organizationId: string }): Promise<{
      data: { message: string } | null;
      error: ApiError | null;
    }> {
      // Client-side only: the active workspace is a URL/localStorage concern
      // in the Go backend world, no server round-trip.
      localStorage.setItem("opus.active_workspace", input.organizationId);
      return { data: { message: "ok" }, error: null };
    },

    // Client-side role checks (no server round-trip). The Go workspace module
    // owns access enforcement; these answer the same question from the member
    // role the UI already holds, for rendering controls.
    async hasPermission(input: {
      permissions?: Record<string, string[]>;
      role?: string;
    }) {
      const role = input.role as WorkspaceMember["role"];
      if (!role) {
        return { data: false, error: null };
      }
      return { data: hasRolePermission(role, input.permissions), error: null };
    },

    checkRolePermission(input: {
      permissions?: Record<string, string[]>;
      role: string;
    }): boolean {
      return hasRolePermission(
        input.role as WorkspaceMember["role"],
        input.permissions,
      );
    },
  },

  // Session hydration — called by AuthProvider on mount and after login.
  __setSession(user: User | null, isPending = false) {
    setSessionState({ user, isPending });
  },
};
