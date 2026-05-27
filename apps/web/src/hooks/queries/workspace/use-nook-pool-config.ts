import { useQuery } from "@tanstack/react-query";
import getNookPoolConfig from "@/fetchers/nook-pool/get-config";

function useNookPoolConfig(workspaceId: string | undefined) {
  return useQuery({
    queryKey: ["nook-pool-config", workspaceId],
    queryFn: () => getNookPoolConfig(workspaceId as string),
    enabled: !!workspaceId,
  });
}

export default useNookPoolConfig;
