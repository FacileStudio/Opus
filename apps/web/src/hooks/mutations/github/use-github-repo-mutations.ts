import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  linkGitHubRepo,
  unlinkGitHubRepo,
} from "@/fetchers/github/github-repos";

function useInvalidateGitHub(projectId: string) {
  const queryClient = useQueryClient();
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["github-repos", projectId] }),
      queryClient.invalidateQueries({ queryKey: ["github-issues", projectId] }),
    ]);
}

export function useLinkGitHubRepo(projectId: string) {
  const invalidate = useInvalidateGitHub(projectId);
  return useMutation({
    mutationFn: (repo: string) => linkGitHubRepo(projectId, repo),
    onSuccess: invalidate,
  });
}

export function useUnlinkGitHubRepo(projectId: string) {
  const invalidate = useInvalidateGitHub(projectId);
  return useMutation({
    mutationFn: (repoId: number) => unlinkGitHubRepo(projectId, repoId),
    onSuccess: invalidate,
  });
}
