import { useMutation, useQueryClient } from "@tanstack/react-query";
import deleteNookPoolConfig from "@/fetchers/nook-pool/delete-config";

function useDeleteNookPool() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ workspaceId }: { workspaceId: string }) => {
      return deleteNookPoolConfig(workspaceId);
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["nook-pool-config", variables.workspaceId],
      });
    },
  });
}

export default useDeleteNookPool;
