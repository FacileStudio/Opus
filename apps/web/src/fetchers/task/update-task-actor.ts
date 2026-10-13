import { client } from "@opus/libs";
import type Task from "@/types/task";

type UpdateTaskActorPayload = Pick<Task, "userId">;

async function updateTaskActor(taskId: string, task: UpdateTaskActorPayload) {
  const response = await client.task.actor[":id"].$put({
    param: { id: taskId },
    json: {
      userId: task.userId || undefined,
    },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  const data = await response.json();

  return data;
}

export default updateTaskActor;
