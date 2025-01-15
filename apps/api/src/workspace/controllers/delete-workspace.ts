import { and, eq } from "drizzle-orm";
import workspace from "..";
import db from "../../database";
import { workspaceTable } from "../../database/schema";
import type { DeleteWorkspacePayload } from "../db/queries";

async function deleteWorkspace({
  userId,
  workspaceId,
}: DeleteWorkspacePayload) {
  const workspace = await db
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

  const isWorkspaceExisting = Boolean(workspace.at(0));

  if (!isWorkspaceExisting) {
    throw new Error("Workspace not found or access denied");
  }

  const deletedWorkspace = await db
    .delete(workspaceTable)
    .where(eq(workspaceTable.id, workspaceId))
    .returning({
      id: workspaceTable.id,
      name: workspaceTable.name,
      ownerId: workspaceTable.ownerId,
      createdAt: workspaceTable.createdAt,
    });

  return deletedWorkspace.at(0);
}

export default deleteWorkspace;
