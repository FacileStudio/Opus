import { eq } from "drizzle-orm";
import db from "../../database";
import { workspaceTable, workspaceUserTable } from "../../database/schema";

async function getUserWorkspaces(userId: string) {
  const workspaces = await db
    .select({
      id: workspaceTable.id,
      name: workspaceTable.name,
      slug: workspaceTable.slug,
      logo: workspaceTable.logo,
      description: workspaceTable.description,
      createdAt: workspaceTable.createdAt,
    })
    .from(workspaceUserTable)
    .innerJoin(
      workspaceTable,
      eq(workspaceUserTable.workspaceId, workspaceTable.id),
    )
    .where(eq(workspaceUserTable.userId, userId));

  return workspaces;
}

export default getUserWorkspaces;
