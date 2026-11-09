import { useQuery } from "@tanstack/react-query";
import { getTaskPullRequests } from "@/fetchers/github/get-task-pull-requests";

export function useGetTaskPullRequests(projectId: string, taskId: string) {
  return useQuery({
    queryKey: ["task-pull-requests", projectId, taskId],
    queryFn: () => getTaskPullRequests(projectId, taskId),
    enabled: Boolean(projectId && taskId),
  });
}
