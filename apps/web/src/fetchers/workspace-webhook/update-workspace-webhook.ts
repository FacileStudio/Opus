import { getApiUrl } from "@/fetchers/get-api-url";
import type { WorkspaceWebhook } from "./get-workspace-webhooks";

export async function updateWorkspaceWebhook(
  workspaceId: string,
  webhookId: string,
  data: {
    name?: string;
    url?: string;
    secret?: string | null;
    events?: string;
    enabled?: boolean;
  },
): Promise<WorkspaceWebhook> {
  const response = await fetch(
    getApiUrl(`/workspace-webhook/${workspaceId}/${webhookId}`),
    {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    },
  );
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }
  return (await response.json()) as WorkspaceWebhook;
}
