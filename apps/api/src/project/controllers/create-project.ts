import db from "../../database";
import { projectTable } from "../../database/schema";
import type { CreateProjectPayload } from "../db/queries";

function createProject(
  body: Pick<CreateProjectPayload, "name" | "description" | "workspaceId">,
) {
  return db.insert(projectTable).values(body);
}

export default createProject;
