import { eq } from "drizzle-orm";
import db from "../../database";
import {
  userTable,
  workspaceTable,
  workspaceUserTable,
} from "../../database/schema";

function getWorkspaceUsers({ workspaceId }: { workspaceId: string }) {
  return db
    .select({
      userId: userTable.id,
      userName: userTable.name,
    })
    .from(workspaceTable)
    .where(eq(workspaceTable.id, workspaceId))
    .innerJoin(
      workspaceUserTable,
      eq(workspaceTable.id, workspaceUserTable.workspaceId),
    )
    .innerJoin(userTable, eq(workspaceUserTable.userId, userTable.id))
    .unionAll(
      db
        .select({
          userId: userTable.id,
          userName: userTable.name,
        })
        .from(workspaceTable)
        .innerJoin(userTable, eq(workspaceTable.ownerId, userTable.id))
        .where(eq(workspaceTable.id, workspaceId)),
    );
}

export default getWorkspaceUsers;
