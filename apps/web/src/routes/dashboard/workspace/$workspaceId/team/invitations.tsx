import InviteTeamMemberModal from "@/components/team/invite-team-member-modal";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import useGetPendingWorkspaceUsers from "@/hooks/queries/workspace-users/use-get-pending-workspace-users";
import { createFileRoute } from "@tanstack/react-router";
import { UserPlus } from "lucide-react";
import { useState } from "react";

export const Route = createFileRoute(
  "/dashboard/workspace/$workspaceId/team/invitations",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { workspaceId } = Route.useParams();
  const { data: users } = useGetPendingWorkspaceUsers({ workspaceId });
  const [isInviteTeamMemberModalOpen, setIsInviteTeamMemberModalOpen] =
    useState(false);

  return (
    <div className="flex-1 p-4 md:p-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
        <h1 className="text-xl md:text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          Team Invitations
        </h1>
        <Button
          onClick={() => setIsInviteTeamMemberModalOpen(true)}
          className="bg-indigo-600 text-white hover:bg-indigo-500 dark:bg-indigo-500 dark:hover:bg-indigo-400 w-full md:w-auto"
        >
          <UserPlus className="w-4 h-4 mr-2" />
          New Invitation
        </Button>
      </div>

      <div className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
        <div className="overflow-x-auto">
          <div className="md:hidden">
            {users?.map((invitation) => (
              <div
                key={invitation.id}
                className="p-4 border-b border-zinc-200 dark:border-zinc-800 last:border-0"
              >
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback className="bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400 uppercase">
                      {invitation.email[0]}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {invitation.email}
                    </div>
                    <div className="text-xs text-zinc-500 dark:text-zinc-400">
                      Invited{" "}
                      {new Date(invitation.invitedAt).toLocaleDateString(
                        "en-US",
                        {
                          year: "numeric",
                          month: "short",
                          day: "numeric",
                        },
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <table className="w-full hidden md:table">
            <thead>
              <tr className="border-b border-zinc-200 dark:border-zinc-800">
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Email
                </th>
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Role
                </th>
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Sent
                </th>
              </tr>
            </thead>
            <tbody>
              {users?.map((invitation) => (
                <tr
                  key={invitation.id}
                  className="border-b border-zinc-200 dark:border-zinc-800 last:border-0"
                >
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      <Avatar className="h-6 w-6">
                        <AvatarFallback className="bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400 uppercase text-xs">
                          {invitation.email[0]}
                        </AvatarFallback>
                      </Avatar>
                      <span className="text-sm text-zinc-900 dark:text-zinc-100">
                        {invitation.email}
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400">
                      Admin
                    </span>
                  </td>
                  <td className="py-3 px-4 text-sm text-zinc-500 dark:text-zinc-400">
                    {new Date(invitation.invitedAt).toLocaleDateString(
                      "en-US",
                      {
                        year: "numeric",
                        month: "short",
                        day: "numeric",
                      },
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {users?.length === 0 && (
          <div className="py-12 text-center">
            <p className="text-zinc-500 dark:text-zinc-400">
              No invitations found
            </p>
          </div>
        )}
      </div>

      <InviteTeamMemberModal
        open={isInviteTeamMemberModalOpen}
        onClose={() => setIsInviteTeamMemberModalOpen(false)}
      />
    </div>
  );
}
