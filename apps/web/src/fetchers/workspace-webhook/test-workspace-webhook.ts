import { getApiUrl } from "@/fetchers/get-api-url";

export type WebhookTestResult = {
  success: boolean;
  statusCode?: number;
  error?: string;
};

export async function testWorkspaceWebhook(
  workspaceId: string,
  webhookId: string,
): Promise<WebhookTestResult> {
  const response = await fetch(
    getApiUrl(`/workspace-webhook/${workspaceId}/${webhookId}/test`),
    {
      method: "POST",
      credentials: "include",
    },
  );
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }
  return (await response.json()) as WebhookTestResult;
}
