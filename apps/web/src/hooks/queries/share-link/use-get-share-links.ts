import { useQuery } from "@tanstack/react-query";
import getShareLinks from "@/fetchers/share-link/get-share-links";

function useGetShareLinks(projectId: string) {
  return useQuery({
    queryKey: ["share-links", projectId],
    queryFn: () => getShareLinks(projectId),
    enabled: !!projectId,
  });
}

export default useGetShareLinks;
