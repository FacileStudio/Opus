import { authClient } from "@/lib/auth-client";

export type GetWorkspaceUsersRequest = {
  workspaceId: string;
  limit?: number;
  offset?: number;
  sortBy?: string;
  sortDirection?: "asc" | "desc";
};

async function getWorkspaceUsers({
  workspaceId,
  limit,
  offset,
}: GetWorkspaceUsersRequest) {
  const { data, error } = await authClient.organization.listMembers({
    query: {
      organizationId: workspaceId,
      limit,
      offset,
    },
  });

  if (error) {
    throw new Error(error.message || "Failed to fetch space users");
  }

  return data || { members: [] };
}

export default getWorkspaceUsers;
