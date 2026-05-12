import { useQuery } from "@tanstack/react-query";
import { getWorkspaceWebhooks } from "@/fetchers/workspace-webhook/get-workspace-webhooks";

export function useGetWorkspaceWebhooks(workspaceId: string) {
  return useQuery({
    queryKey: ["workspace-webhooks", workspaceId],
    queryFn: () => getWorkspaceWebhooks(workspaceId),
    enabled: Boolean(workspaceId),
  });
}
