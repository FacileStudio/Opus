import { useMutation, useQueryClient } from "@tanstack/react-query";
import createShareLink from "@/fetchers/share-link/create-share-link";

function useCreateShareLink(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createShareLink,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["share-links", projectId] });
    },
  });
}

export default useCreateShareLink;
