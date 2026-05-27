import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { SaveNookPoolConfigRequest } from "@/fetchers/nook-pool/save-config";
import saveNookPoolConfig from "@/fetchers/nook-pool/save-config";

function useSaveNookPool() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      workspaceId,
      ...data
    }: SaveNookPoolConfigRequest & { workspaceId: string }) => {
      return saveNookPoolConfig(workspaceId, data);
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["nook-pool-config", variables.workspaceId],
      });
    },
  });
}

export default useSaveNookPool;
