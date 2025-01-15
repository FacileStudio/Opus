import { eq, or } from "drizzle-orm";
import db from "../../database";
import { workspaceTable, workspaceUserTable } from "../../database/schema";

// TODO: Unify this type
async function getWorkspaces({ userId }: { userId: string }) {
  return await db
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
    );
}

export default getWorkspaces;
