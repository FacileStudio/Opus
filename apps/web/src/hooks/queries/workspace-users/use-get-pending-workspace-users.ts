import getPendingWorkspaceUsers from "@/fetchers/workspace-user/get-pending-workspace-users";
import { useQuery } from "@tanstack/react-query";

function useGetPendingWorkspaceUsers({ workspaceId }: { workspaceId: string }) {
  return useQuery({
    queryKey: ["pending-workspace-users", workspaceId],
    queryFn: () => getPendingWorkspaceUsers({ workspaceId }),
  });
}

export default useGetPendingWorkspaceUsers;
