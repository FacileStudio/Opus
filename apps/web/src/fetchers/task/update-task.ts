import { client } from "@opus/libs";
import type Task from "@/types/task";

async function updateTask(taskId: string, task: Task) {
  const response = await client.task[":id"].$put({
    param: { id: taskId },
    json: {
      userId: task.userId || undefined,
      title: task.title,
      description: task.description || "",
      status: task.status,
      priority: task.priority || "no-priority",
      startDate: task.startDate ?? null,
      dueDate: task.dueDate ?? null,
      position: task.position ?? 0,
      projectId: task.projectId,
    },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  const data = await response.json();

  return data;
}

export default updateTask;
