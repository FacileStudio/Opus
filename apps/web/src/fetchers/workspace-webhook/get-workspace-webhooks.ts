import { getApiUrl } from "@/fetchers/get-api-url";

export type WorkspaceWebhook = {
  id: string;
  workspaceId: string;
  name: string;
  url: string;
  events: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
};

export async function getWorkspaceWebhooks(
  workspaceId: string,
): Promise<WorkspaceWebhook[]> {
  const response = await fetch(getApiUrl(`/workspace-webhook/${workspaceId}`), {
    credentials: "include",
  });
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }
  return (await response.json()) as WorkspaceWebhook[];
}
