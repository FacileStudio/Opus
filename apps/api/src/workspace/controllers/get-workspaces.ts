import { eq, or } from "drizzle-orm";
import db from "../../database";
import {
  projectTable,
  workspaceTable,
  workspaceUserTable,
} from "../../database/schema";
import groupWorkspacesWithProjects from "../utils/group-workspaces-with-project";

async function getWorkspaces({ userId }: { userId: string }) {
  const workspacesWithProjects = await db
    .select({
      id: workspaceTable.id,
      name: workspaceTable.name,
      ownerId: workspaceTable.ownerId,
      projectId: projectTable.id,
      projectName: projectTable.name,
    })
    .from(workspaceTable)
    .leftJoin(
      workspaceUserTable,
      eq(workspaceTable.id, workspaceUserTable.workspaceId),
    )
    .leftJoin(projectTable, eq(workspaceTable.id, projectTable.workspaceId))
    .where(
      or(
        eq(workspaceTable.ownerId, userId),
        eq(workspaceUserTable.userId, userId),
      ),
    );

  return groupWorkspacesWithProjects(workspacesWithProjects);
}

export default getWorkspaces;
