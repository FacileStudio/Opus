import { useMutation, useQueryClient } from "@tanstack/react-query";
import deleteShareLink from "@/fetchers/share-link/delete-share-link";

function useDeleteShareLink(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteShareLink,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["share-links", projectId] });
    },
  });
}

export default useDeleteShareLink;
