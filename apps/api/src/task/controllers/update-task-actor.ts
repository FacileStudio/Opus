import { eq } from "drizzle-orm";
import { HTTPException } from "hono/http-exception";
import db from "../../database";
import { taskTable, userTable } from "../../database/schema";
import { publishEvent } from "../../events";

async function updateTaskActor({
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

  const nextActorId = userId || null;
  if (existingTask.actorId === nextActorId) {
    return existingTask;
  }

  const [updatedTask] = await db
    .update(taskTable)
    .set({ actorId: nextActorId || null })
    .where(eq(taskTable.id, id))
    .returning();

  if (!updatedTask) {
    throw new HTTPException(500, {
      message: "Failed to update task actor",
    });
  }

  const [newActorName, oldActorName] = await Promise.all([
    userId
      ? db
          .select({ name: userTable.name })
          .from(userTable)
          .where(eq(userTable.id, userId))
          .limit(1)
          .then((rows) => rows[0]?.name)
      : undefined,
    existingTask.actorId
      ? db
          .select({ name: userTable.name })
          .from(userTable)
          .where(eq(userTable.id, existingTask.actorId))
          .limit(1)
          .then((rows) => rows[0]?.name)
      : undefined,
  ]);

  if (!userId) {
    await publishEvent("task.unassigned", {
      taskId: updatedTask.id,
      projectId: updatedTask.projectId,
      userId: currentUserId,
      oldActor: oldActorName,
      title: updatedTask.title,
      type: "unassigned",
    });

    return updatedTask;
  }

  await publishEvent("task.actor_changed", {
    taskId: updatedTask.id,
    projectId: updatedTask.projectId,
    userId: currentUserId,
    oldActor: oldActorName,
    newActor: newActorName,
    newActorId: userId,
    title: updatedTask.title,
    type: "actor_changed",
  });

  return updatedTask;
}

export default updateTaskActor;
