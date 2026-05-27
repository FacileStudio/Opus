import { getApiUrl } from "@/fetchers/get-api-url";

async function deleteNookPoolConfig(workspaceId: string) {
  const response = await fetch(getApiUrl(`/nook-pool/${workspaceId}`), {
    method: "DELETE",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json() as Promise<{ success: boolean }>;
}

export default deleteNookPoolConfig;
