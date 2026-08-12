import { authClient } from "@/lib/auth-client";
import type { WorkspaceUserInvitation } from "@/types/workspace-user";

export async function getPendingInvitations(): Promise<
  WorkspaceUserInvitation[]
> {
  const { data, error } = await authClient.organization.listUserInvitations();

  if (error) {
    throw new Error(error.message || "Failed to get pending invitations");
  }

  return (data || []).filter(
    (invitation) => invitation.status === "pending",
  ) as WorkspaceUserInvitation[];
}
