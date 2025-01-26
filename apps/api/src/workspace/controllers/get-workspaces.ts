import { eq, or } from "drizzle-orm";
import db from "../../database";
import {
  projectTable,
  workspaceTable,
  workspaceUserTable,
} from "../../database/schema";

async function getWorkspaces({ userId }: { userId: string }) {
  const workspaces = await db
    .select({
      id: workspaceTable.id,
      name: workspaceTable.name,
      ownerId: workspaceTable.ownerId,
    })
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
    )
    .groupBy(workspaceTable.id, workspaceTable.name, workspaceTable.ownerId);

  const workspacesWithProjects = await Promise.all(
    workspaces.map(async (workspace) => ({
      ...workspace,
      projects: await db
        .select({
          id: projectTable.id,
          name: projectTable.name,
        })
        .from(projectTable)
        .where(eq(projectTable.workspaceId, workspace.id)),
    })),
  );

  return workspacesWithProjects;
}

export default getWorkspaces;
