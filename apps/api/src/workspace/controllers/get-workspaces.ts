import db from "../../database";
import { workspaceTable } from "../../database/schema";

async function getWorkspaces() {
  return db.select().from(workspaceTable);
}

export default getWorkspaces;
