import db from "../../database";
import { workspaceTable } from "../../database/schema";
import type { CreateWorkspacePayload } from "../db/queries";

async function createWorkspace(body: CreateWorkspacePayload) {
  return db.insert(workspaceTable).values(body);
}

export default createWorkspace;
