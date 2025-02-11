import Elysia, { t } from "elysia";
import getPendingWorkspaceUsers from "./controllers/get-pending-workspace-users";
import getWorkspaceUsers from "./controllers/get-workspace-users";
import inviteWorkspaceUser from "./controllers/invite-workspace-user";

const workspaceUser = new Elysia({ prefix: "/workspace-user" })
  .get("/list/:workspaceId", async ({ params: { workspaceId } }) => {
    const workspaceUsersInWorkspace = await getWorkspaceUsers({ workspaceId });

    return workspaceUsersInWorkspace;
  })
  .get("/pending/list/:workspaceId", async ({ params: { workspaceId } }) => {
    const pendingWorkspaceUsersInWorkspace = await getPendingWorkspaceUsers({
      workspaceId,
    });

    return pendingWorkspaceUsersInWorkspace;
  })
  .post(
    "/:workspaceId/invite",
    async ({ body }) => {
      const invitedWorkspaceUser = await inviteWorkspaceUser(body);

      return invitedWorkspaceUser;
    },
    {
      body: t.Object({
        userEmail: t.String(),
        workspaceId: t.String(),
      }),
    },
  );

export default workspaceUser;
