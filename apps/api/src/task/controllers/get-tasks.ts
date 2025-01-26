import { eq } from "drizzle-orm";
import db from "../../database";
import { projectTable, taskTable, userTable } from "../../database/schema";

const DEFAULT_COLUMNS = [
  { id: "to-do", name: "To Do" },
  { id: "in-progress", name: "In Progress" },
  { id: "in-review", name: "In Review" },
  { id: "done", name: "Done" },
] as const;

async function getTasks(projectId: string) {
  const project = await db.query.projectTable.findFirst({
    where: eq(projectTable.id, projectId),
  });

  if (!project) {
    throw new Error("Project not found");
  }

  const tasks = await db
    .select({
      id: taskTable.id,
      title: taskTable.title,
      description: taskTable.description,
      status: taskTable.status,
      priority: taskTable.priority,
      dueDate: taskTable.dueDate,
      createdAt: taskTable.createdAt,
      assigneeId: taskTable.assigneeId,
      assigneeName: userTable.name,
      assigneeEmail: userTable.email,
    })
    .from(taskTable)
    .leftJoin(userTable, eq(taskTable.assigneeId, userTable.id))
    .where(eq(taskTable.projectId, projectId));

  const columns = DEFAULT_COLUMNS.map((column) => ({
    id: column.id,
    name: column.name,
    tasks: tasks
      .filter((task) => task.status === column.id)
      .map((task) => ({
        ...task,
      })),
  }));

  return {
    id: project.id,
    name: project.name,
    workspaceId: project.workspaceId,
    columns,
  };
}

export default getTasks;
