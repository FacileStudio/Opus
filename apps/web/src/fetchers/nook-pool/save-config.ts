import { getApiUrl } from "@/fetchers/get-api-url";
import type { NookPoolConfig } from "./get-config";

export type SaveNookPoolConfigRequest = {
  instanceUrl: string;
  secret: string;
  enabled?: boolean;
};

async function saveNookPoolConfig(
  workspaceId: string,
  json: SaveNookPoolConfigRequest,
) {
  const response = await fetch(getApiUrl(`/nook-pool/${workspaceId}`), {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(json),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return (await response.json()) as NonNullable<NookPoolConfig>;
}

export default saveNookPoolConfig;
