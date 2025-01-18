import { api } from "@kaneo/libs";

async function getProjects({ workspaceId }: { workspaceId: string }) {
  const response = await api.project.list({ workspaceId }).get();

  if (response.error) {
    throw new Error(response.error.value.message);
  }

  return response?.data;
}

export default getProjects;
