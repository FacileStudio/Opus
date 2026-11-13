import { apiFetch, getApiUrl } from "@/fetchers/get-api-url";

type TriggerSyncResult = {
  success: boolean;
  projects_synced?: number;
  tasks_synced?: number;
  error?: string;
};

async function triggerNookPoolSync(workspaceId: string) {
  const response = await apiFetch(getApiUrl(`/nook-pool/${workspaceId}/sync`), {
    method: "POST",
    credentials: "include",
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return (await response.json()) as TriggerSyncResult;
}

export default triggerNookPoolSync;
