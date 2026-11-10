import { useQuery } from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import { authClient } from "@/lib/auth-client";

function useActiveWorkspace() {
  const {
    data: organizations,
    error,
    isPending,
  } = useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await authClient.organization.list();
      if (error) {
        throw new Error(error.message || "Failed to fetch workspaces");
      }
      return data || [];
    },
  });

  const { workspaceId } = useParams({
    strict: false,
    select: (params) => ({
      workspaceId:
        "workspaceId" in params && typeof params.workspaceId === "string"
          ? params.workspaceId
          : undefined,
    }),
  });

  const activeId =
    workspaceId ?? localStorage.getItem("opus.active_workspace") ?? undefined;
  const workspace = activeId
    ? organizations?.find((organization) => organization.id === activeId)
    : undefined;
  const isLoading =
    (!!activeId && isPending && !workspace) || (!workspace && !error);

  return {
    data: workspace,
    error,
    isLoading,
    isError: !!error,
    isSettled: !isPending,
  };
}

export default useActiveWorkspace;
