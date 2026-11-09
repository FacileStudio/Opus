import { getApiUrl } from "@/fetchers/get-api-url";
import type { PullRequest } from "@/types/pull-request";

export async function getTaskPullRequests(
  projectId: string,
  taskId: string,
): Promise<PullRequest[]> {
  if (!taskId || !projectId) return [];
  const response = await fetch(
    getApiUrl(`/project/${projectId}/task/${taskId}/pull-requests/`),
    {
      credentials: "include",
      headers: {
        Accept: "application/json",
      },
    },
  );
  if (!response.ok) {
    return [];
  }
  return response.json();
}
