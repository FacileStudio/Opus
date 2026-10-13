import { UserPlus } from "@solar-icons/react";
import {
  createFileRoute,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import type { ReactNode } from "react";
import { createContext, useContext, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";
import InviteTeamMemberModal from "@/components/team/invite-team-member-modal";
import MembersTable from "@/components/team/members-table";
import { Button } from "@/components/ui/button";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import useGetFullWorkspace from "@/hooks/queries/workspace/use-get-full-workspace";

const workspaceMembersPath =
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/members";

const WorkspaceMembersContent = createContext<ReactNode>(null);

function WorkspaceMembersInnerRoute() {
  return <>{useContext(WorkspaceMembersContent)}</>;
}

function buildWorkspaceMembersRouter(workspaceId: string) {
  const rootRoute = createRootRoute();
  const membersRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: workspaceMembersPath,
    component: WorkspaceMembersInnerRoute,
  });

  return createRouter({
    routeTree: rootRoute.addChildren([membersRoute]),
    history: createMemoryHistory({
      initialEntries: [
        workspaceMembersPath.replace("$workspaceId", workspaceId),
      ],
    }),
  });
}

type WorkspaceMembersSectionProps = {
  workspaceId: string;
  isInviteOpen: boolean;
  onCloseInvite: () => void;
};

function WorkspaceMembersSection({
  workspaceId,
  isInviteOpen,
  onCloseInvite,
}: WorkspaceMembersSectionProps) {
  const { data: workspace } = useGetFullWorkspace({ workspaceId });
  const router = useMemo(
    () => buildWorkspaceMembersRouter(workspaceId),
    [workspaceId],
  );

  const content = (
    <>
      <MembersTable
        users={workspace?.members ?? []}
        invitations={workspace?.invitations ?? []}
      />
      <InviteTeamMemberModal open={isInviteOpen} onClose={onCloseInvite} />
    </>
  );

  return (
    <WorkspaceMembersContent.Provider value={content}>
      <RouterProvider router={router} />
    </WorkspaceMembersContent.Provider>
  );
}

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace/members",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();
  const { data: activeWorkspace, isLoading } = useActiveWorkspace();
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const workspaceId = activeWorkspace?.id;

  return (
    <>
      <PageTitle title={t("team:members.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-xl font-semibold">
            {t("team:members.pageTitle")}
          </h1>
          {!!workspaceId && !isLoading && (
            <Button
              onClick={() => setIsInviteOpen(true)}
              variant="outline"
              size="xs"
              className="gap-1 w-full md:w-auto"
            >
              <UserPlus className="w-3 h-3" />
              {t("team:members.inviteMember")}
            </Button>
          )}
        </div>

        {isLoading ? (
          <div className="flex min-h-40 items-center justify-center">
            <p className="text-sm text-muted-foreground">
              {t("common:empty.loading")}
            </p>
          </div>
        ) : workspaceId ? (
          <WorkspaceMembersSection
            workspaceId={workspaceId}
            isInviteOpen={isInviteOpen}
            onCloseInvite={() => setIsInviteOpen(false)}
          />
        ) : null}
      </div>
    </>
  );
}
