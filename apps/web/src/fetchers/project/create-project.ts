import { api } from "@kaneo/libs";

async function createProject({
  name,
  description,
  workspaceId,
  icon,
}: { name: string; description: string; workspaceId: string; icon: string }) {
  const response = await api.project.create.post({
    name,
    description,
    workspaceId,
    icon,
  });

  if (response.error) {
    throw new Error(response.error.value.message);
  }

  return response;
}

export default createProject;
