import { api } from "@kaneo/libs";

async function createProject({
  name,
  description,
  workspaceId,
}: { name: string; description: string; workspaceId: string }) {
  const response = await api.project.create.post({
    name,
    description,
    workspaceId,
  });

  if (response.error) {
    throw new Error(response.error.value.message);
  }

  return response;
}

export default createProject;
