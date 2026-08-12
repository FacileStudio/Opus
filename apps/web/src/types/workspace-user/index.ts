import type { WorkspaceMember, WorkspaceInvitation } from "@/lib/auth-client";

export type WorkspaceUser = WorkspaceMember;

export type ActiveWorkspaceUser = WorkspaceMember;

export type WorkspaceUserInvitation = WorkspaceInvitation;

export type UserInvitation = WorkspaceInvitation;

export default WorkspaceUser;