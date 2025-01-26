import { eq } from "drizzle-orm";
import db from "../../database";
import { taskTable, userTable } from "../../database/schema";

async function createTask(body: {
  projectId: string;
  assigneeId: string;
  title: string;
  status: string;
  dueDate: Date | null;
  description: string;
  priority: string;
}) {
  const [assignee] = await db
    .select({ name: userTable.name })
    .from(userTable)
    .where(eq(userTable.id, body.assigneeId));

  if (!assignee) {
    throw new Error("Assignee not found");
  }

  const [createdTask] = await db.insert(taskTable).values(body).returning();

  return {
    ...createdTask,
    assigneeName: assignee.name,
  };
}

export default createTask;
