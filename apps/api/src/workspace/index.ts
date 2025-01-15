import Elysia, { t } from "elysia";
import createWorkspace from "./controllers/create-workspace";
import deleteWorkspace from "./controllers/delete-workspace";
import getWorkspace from "./controllers/get-workspace";
import getWorkspaces from "./controllers/get-workspaces";
import updateWorkspace from "./controllers/update-workspace";
import { updateWorkspaceSchema } from "./db/queries";

const workspace = new Elysia({ prefix: "/workspace" })
  .state("userId", "")
  .post(
    "/create",
    async ({ body, store }) => {
      const userId = store.userId;

      const createdWorkspace = await createWorkspace({
        ...body,
        ownerId: userId,
      });

      return createdWorkspace;
    },
    {
      body: t.Object({
        name: t.String(),
      }),
    },
  )
  .get("/list", async ({ store }) => {
    const userId = store.userId;

    const workspaces = await getWorkspaces({ userId });

    return workspaces;
  })
  .get("/:id", async ({ store, params }) => {
    const userId = store.userId;
    const workspaceId = params.id;

    const workspace = await getWorkspace({ userId, workspaceId });

    return workspace;
  })
  .put(
    "/:id",
    async ({ store, params, body }) => {
      const userId = store.userId;
      const workspaceId = params.id;

      const updatedWorkspace = await updateWorkspace({
        userId,
        workspaceId,
        body,
      });

      return updatedWorkspace;
    },
    { body: updateWorkspaceSchema },
  )
  .delete("/:id", async ({ store, params }) => {
    const userId = store.userId;
    const workspaceId = params.id;

    const deletedWorkspace = await deleteWorkspace({ userId, workspaceId });

    return deletedWorkspace;
  });

export default workspace;
