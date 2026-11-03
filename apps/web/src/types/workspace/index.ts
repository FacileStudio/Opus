import type { WorkspaceMember } from "@/lib/auth-client";

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  logo: string | null;
  description: string | null;
  createdAt: string;
  registreSpaceId: string | null;
};

export type FullWorkspace = Workspace & {
  members: WorkspaceMember[];
  invitations: WorkspaceInvitation[];
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

export default Workspace;
