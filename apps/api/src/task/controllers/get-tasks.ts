import { eq } from "drizzle-orm";
import db from "../../database";
import { projectTable, taskTable } from "../../database/schema";

const DEFAULT_COLUMNS = [
  { id: "to-do", name: "To Do" },
  { id: "in-progress", name: "In Progress" },
  { id: "done", name: "Done" },
] as const;

async function getTasks(projectId: string) {
  const project = await db.query.projectTable.findFirst({
    where: eq(projectTable.id, projectId),
  });

  if (!project) {
    throw new Error("TODO");
  }

  const tasks = await db.query.taskTable.findMany({
    where: eq(taskTable.projectId, projectId),
    orderBy: (tasks, { desc }) => [desc(tasks.createdAt)],
  });

  const statuses = [...new Set(tasks.map((task) => task.status))];

  const columns = DEFAULT_COLUMNS.map((column) => ({
    id: column.id,
    name: column.name,
    tasks: tasks.filter((task) => task.status === column.id),
  }));

  return {
    id: project.id,
    name: project.name,
    workspaceId: project.workspaceId,
    columns,
  };
}

export default getTasks;
