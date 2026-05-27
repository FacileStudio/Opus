import { getApiUrl } from "@/fetchers/get-api-url";

export type NookPoolConfig = {
  id: string;
  workspaceId: string;
  instanceUrl: string;
  enabled: boolean;
  connected: boolean;
  fromEnv?: boolean;
  createdAt: string;
  updatedAt: string;
} | null;

async function getNookPoolConfig(workspaceId: string) {
  const response = await fetch(getApiUrl(`/nook-pool/${workspaceId}`), {
    credentials: "include",
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return (await response.json()) as NookPoolConfig;
}

export default getNookPoolConfig;
