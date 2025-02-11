import { api } from "@kaneo/libs";

async function getPendingWorkspaceUsers({
  workspaceId,
}: { workspaceId: string }) {
  const response = await api["workspace-user"].pending
    .list({ workspaceId })
    .get();

  if (response.error) {
    throw new Error(response.error.value.message);
  }

  return response.data;
}

export default getPendingWorkspaceUsers;
