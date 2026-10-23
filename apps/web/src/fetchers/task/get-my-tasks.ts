import { client } from "@opus/libs";

async function getMyTasks(workspaceId: string) {
  const response = await client.task.me.$get({ query: { workspaceId } });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

export default getMyTasks;
