import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "@/lib/auth-client";
import type { IdeaRun, IdeasPageData } from "@/lib/ideas";
import { isRunSettled } from "@/lib/ideas";

const POLL_MS = 3000;

export function ideasQueryKey(workspaceId: string | undefined) {
  return ["ideas", workspaceId ?? "personal"] as const;
}

export function ideasPath(path: string, workspaceId: string | undefined) {
  const query = workspaceId
    ? `?workspaceId=${encodeURIComponent(workspaceId)}`
    : "";
  return `/ideas${path}${query}`;
}

export function useIdeas(workspaceId: string | undefined, enabled = true) {
  return useQuery({
    queryKey: ideasQueryKey(workspaceId),
    enabled,
    queryFn: async () => {
      const result = await apiRequest<IdeasPageData>(
        ideasPath("", workspaceId),
      );
      if (result.error) throw new Error(result.error.message);
      return { ...result.data, ideas: result.data.ideas ?? [] };
    },
    refetchInterval: (query) =>
      query.state.data?.generating ? POLL_MS : false,
  });
}

export function useIdeaRun(runId: string | undefined) {
  return useQuery({
    queryKey: ["ideas", "run", runId],
    enabled: !!runId,
    queryFn: async () => {
      const result = await apiRequest<IdeaRun>(`/ideas/runs/${runId}`);
      if (result.error) throw new Error(result.error.message);
      return result.data;
    },
    refetchInterval: (query) =>
      isRunSettled(query.state.data?.state) ? false : 2000,
  });
}
