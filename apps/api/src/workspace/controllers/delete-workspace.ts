import { and, eq } from "drizzle-orm";
import db from "../../database";
import { workspaceTable } from "../../database/schema";
import type { DeleteWorkspacePayload } from "../db/queries";

async function deleteWorkspace({
  userId,
  workspaceId,
}: DeleteWorkspacePayload) {
  const [existingWorkspace] = await db
    .select({
      id: workspaceTable.id,
      ownerId: workspaceTable.ownerId,
    })
    .from(workspaceTable)
    .where(
      and(
        eq(workspaceTable.id, workspaceId),
        eq(workspaceTable.ownerId, userId),
      ),
    )
    .limit(1);

  const isWorkspaceExisting = Boolean(existingWorkspace);

  if (!isWorkspaceExisting) {
    throw new Error("Workspace not found or access denied");
  }

  const [deletedWorkspace] = await db
    .delete(workspaceTable)
    .where(eq(workspaceTable.id, workspaceId))
    .returning({
      id: workspaceTable.id,
      name: workspaceTable.name,
      ownerId: workspaceTable.ownerId,
      createdAt: workspaceTable.createdAt,
    });

  return deletedWorkspace;
}

export default deleteWorkspace;
