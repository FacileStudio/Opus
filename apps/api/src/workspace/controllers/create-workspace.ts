import type { Static } from "elysia";
import db from "../../database";
import { workspaceTable } from "../../database/schema";
import type { createWorkspaceSchema } from "../db/queries";

type CreateWorkspaceArgs = Static<typeof createWorkspaceSchema>;

async function createWorkspace(body: CreateWorkspaceArgs) {
  return db.insert(workspaceTable).values(body);
}

export default createWorkspace;
