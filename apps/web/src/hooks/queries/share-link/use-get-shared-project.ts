import { useQuery } from "@tanstack/react-query";
import getSharedProject from "@/fetchers/share-link/get-shared-project";

function useGetSharedProject(token: string) {
  return useQuery({
    queryKey: ["shared-project", token],
    queryFn: () => getSharedProject(token),
    enabled: !!token,
  });
}

export default useGetSharedProject;
