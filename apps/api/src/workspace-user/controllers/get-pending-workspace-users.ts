import { and, eq, isNull } from "drizzle-orm";
import db from "../../database";
import { userTable, workspaceUserTable } from "../../database/schema";

async function getPendingWorkspaceUsers({
  workspaceId,
}: {
  workspaceId: string;
}) {
  const pendingInvites = await db
    .select({
      id: workspaceUserTable.id,
      email: workspaceUserTable.userEmail,
      role: workspaceUserTable.role,
      invitedAt: workspaceUserTable.joinedAt,
    })
    .from(workspaceUserTable)
    .leftJoin(userTable, eq(workspaceUserTable.userEmail, userTable.email))
    .where(
      and(
        eq(workspaceUserTable.workspaceId, workspaceId),
        isNull(userTable.email),
      ),
    );

  return pendingInvites;
}

export default getPendingWorkspaceUsers;
