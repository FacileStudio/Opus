import { getApiUrl } from "@/fetchers/get-api-url";

export async function deleteWorkspaceWebhook(
  workspaceId: string,
  webhookId: string,
): Promise<void> {
  const response = await fetch(
    getApiUrl(`/workspace-webhook/${workspaceId}/${webhookId}`),
    {
      method: "DELETE",
      credentials: "include",
    },
  );
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }
}
