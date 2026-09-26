import { useQuery } from "@tanstack/react-query";
import { authClient } from "@/lib/auth-client";

function useGetWorkspaces() {
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

  return {
    data: organizations,
    error,
    isLoading: isPending,
    isError: !!error,
  };
}

export default useGetWorkspaces;
