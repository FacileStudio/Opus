import { useMutation, useQueryClient } from "@tanstack/react-query";
import updateShareLink from "@/fetchers/share-link/update-share-link";

function useUpdateShareLink(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: {
        label?: string | null;
        expiresAt?: string | null;
        isActive?: boolean;
      };
    }) => updateShareLink(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["share-links", projectId] });
    },
  });
}

export default useUpdateShareLink;
