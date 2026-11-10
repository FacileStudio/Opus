import { createFileRoute, redirect } from "@tanstack/react-router";
import { getPendingInvitations } from "@/fetchers/invitation/get-pending-invitations";
import getWorkspaces from "@/fetchers/workspace/get-workspaces";
import { authClient } from "@/lib/auth-client";
import { takeReturn } from "@/lib/pending";
import type Workspace from "@/types/workspace";

export const Route = createFileRoute("/_layout/_authenticated/dashboard/")({
  beforeLoad: async () => {
    const back = takeReturn();
    if (back) {
      throw redirect({ href: back });
    }
    const workspaces: Workspace[] = await getWorkspaces();
    const invitations = await getPendingInvitations();

    if (invitations && invitations.length > 0 && !workspaces.length) {
      throw redirect({ to: "/invitations" });
    }

    const activeWorkspaceId = localStorage.getItem("opus.active_workspace");

    if (workspaces && workspaces.length > 0) {
      if (
        activeWorkspaceId &&
        workspaces.some((ws) => ws.id === activeWorkspaceId)
      ) {
        throw redirect({
          to: "/dashboard/workspace/$workspaceId",
          params: { workspaceId: activeWorkspaceId },
        });
      }

      const firstWorkspace = workspaces[0];

      authClient.organization.setActive({
        organizationId: firstWorkspace.id,
      });

      throw redirect({
        to: "/dashboard/workspace/$workspaceId",
        params: { workspaceId: firstWorkspace.id },
      });
    }
    throw redirect({ to: "/onboarding" });
  },
});
