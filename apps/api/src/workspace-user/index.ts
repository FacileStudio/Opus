import Elysia from "elysia";
import getWorkspaceUsers from "./controllers/get-workspace-users";

const workspaceUser = new Elysia({ prefix: "/workspace-user" }).get(
  "/list/:workspaceId",
  async ({ params: { workspaceId } }) => {
    const workspaceUsersInWorkspace = await getWorkspaceUsers({ workspaceId });

    return workspaceUsersInWorkspace;
  },
);

export default workspaceUser;
