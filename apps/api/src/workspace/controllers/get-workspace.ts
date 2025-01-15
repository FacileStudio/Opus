import { and, eq, or } from "drizzle-orm";
import db from "../../database";
import { workspaceTable, workspaceUserTable } from "../../database/schema";

// TODO: Unify this type
async function getWorkspace({
  userId,
  workspaceId,
}: { userId: string; workspaceId: string }) {
  const workspace = await db
    .select({
      id: workspaceTable.id,
      name: workspaceTable.name,
      ownerId: workspaceTable.ownerId,
      createdAt: workspaceTable.createdAt,
    })
    .from(workspaceTable)
    .leftJoin(
      workspaceUserTable,
      eq(workspaceTable.id, workspaceUserTable.workspaceId),
    )
    .where(
      and(
        eq(workspaceTable.id, workspaceId),
        or(
          eq(workspaceTable.ownerId, userId),
          eq(workspaceUserTable.userId, userId),
        ),
      ),
    )
    .limit(1);

  return workspace.at(0);
}

export default getWorkspace;
