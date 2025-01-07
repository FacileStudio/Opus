import { eq, or } from "drizzle-orm";
import Elysia from "elysia";
import db from "../database";
import { workspaceTable, workspaceUserTable } from "../database/schema";
import createWorkspace from "./controllers/create-workspace";
import { createWorkspaceSchema } from "./db/queries";

const workspace = new Elysia({ prefix: "/workspace" })
  .state("userId", "")
  .post(
    "/",
    async ({ body }) => {
      const createdWorkspace = await createWorkspace(body);

      return createdWorkspace;
    },
    {
      body: createWorkspaceSchema,
    },
  )
  .get("/workspaces", async ({ store }) => {
    const userId = store.userId;

    const workspaces = await db
      .select()
      .from(workspaceTable)
      .leftJoin(
        workspaceUserTable,
        eq(workspaceTable.id, workspaceUserTable.workspaceId),
      )
      .where(
        or(
          eq(workspaceTable.ownerId, userId),
          eq(workspaceUserTable.userId, userId),
        ),
      );

    return workspaces;
  })
  .put("/:id", () => {})
  .delete("/:id", () => {});

export default workspace;
