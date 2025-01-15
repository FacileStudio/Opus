import getWorkspaces from "@/fetchers/workspace/get-workspaces";
import { useQuery } from "@tanstack/react-query";

function useGetWorkspaces() {
  return useQuery({
    queryFn: () => getWorkspaces(),
    queryKey: ["workspaces"],
  });
}

export default useGetWorkspaces;
