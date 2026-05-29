import { getApiUrl } from "@/fetchers/get-api-url";

export type PoolEventToggle = {
  event: string;
  enabled: boolean;
};

async function getPoolEvents(workspaceId: string) {
  const response = await fetch(getApiUrl(`/nook-pool/${workspaceId}/events`), {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error(await response.text());
  }

  return (await response.json()) as PoolEventToggle[];
}

export default getPoolEvents;
