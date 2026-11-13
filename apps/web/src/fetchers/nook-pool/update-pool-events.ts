import { apiFetch, getApiUrl } from "@/fetchers/get-api-url";
import type { PoolEventToggle } from "./get-pool-events";

async function updatePoolEvents(
  workspaceId: string,
  events: PoolEventToggle[],
) {
  const response = await apiFetch(
    getApiUrl(`/nook-pool/${workspaceId}/events`),
    {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(events),
    },
  );

  if (!response.ok) {
    throw new Error(await response.text());
  }

  return (await response.json()) as PoolEventToggle[];
}

export default updatePoolEvents;
