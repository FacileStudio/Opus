import { eq } from "drizzle-orm";
import { HTTPException } from "hono/http-exception";
import db from "../../database";
import { taskTable, userTable } from "../../database/schema";
import { publishEvent } from "../../events";

async function updateTaskAssignee({
  id,
  userId,
  currentUserId,
}: {
  id: string;
  userId: string;
  currentUserId: string;
}) {
  const existingTask = await db.query.taskTable.findFirst({
    where: eq(taskTable.id, id),
  });

  if (!existingTask) {
    throw new HTTPException(404, {
      message: "Task not found",
    });
  }

  const nextAssigneeId = userId || null;
  if (existingTask.userId === nextAssigneeId) {
    return existingTask;
  }

  const [updatedTask] = await db
    .update(taskTable)
    .set({ userId: nextAssigneeId || null })
    .where(eq(taskTable.id, id))
    .returning();

  if (!updatedTask) {
    throw new HTTPException(500, {
      message: "Failed to update task assignee",
    });
  }

  const [newAssigneeName, oldAssigneeName] = await Promise.all([
    userId
      ? db
          .select({ name: userTable.name })
          .from(userTable)
          .where(eq(userTable.id, userId))
          .limit(1)
          .then((rows) => rows[0]?.name)
      : undefined,
    existingTask.userId
      ? db
          .select({ name: userTable.name })
          .from(userTable)
          .where(eq(userTable.id, existingTask.userId))
          .limit(1)
          .then((rows) => rows[0]?.name)
      : undefined,
  ]);

  if (!userId) {
    await publishEvent("task.unassigned", {
      taskId: updatedTask.id,
      projectId: updatedTask.projectId,
      userId: currentUserId,
      oldAssignee: oldAssigneeName,
      title: updatedTask.title,
      type: "unassigned",
    });

    return updatedTask;
  }

  await publishEvent("task.assignee_changed", {
    taskId: updatedTask.id,
    projectId: updatedTask.projectId,
    userId: currentUserId,
    oldAssignee: oldAssigneeName,
    newAssignee: newAssigneeName,
    newAssigneeId: userId,
    title: updatedTask.title,
    type: "assignee_changed",
  });

  return updatedTask;
}

export default updateTaskAssignee;
