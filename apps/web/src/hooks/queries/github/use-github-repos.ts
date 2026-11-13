import { useQuery } from "@tanstack/react-query";
import {
  getGitHubIssues,
  getGitHubRepos,
} from "@/fetchers/github/github-repos";
import type { GitHubIssueState } from "@/types/github";

export function useGetGitHubRepos(projectId: string) {
  return useQuery({
    queryKey: ["github-repos", projectId],
    queryFn: () => getGitHubRepos(projectId),
    enabled: Boolean(projectId),
    retry: false,
  });
}

export function useGetGitHubIssues(projectId: string, state: GitHubIssueState) {
  return useQuery({
    queryKey: ["github-issues", projectId, state],
    queryFn: () => getGitHubIssues(projectId, state),
    enabled: Boolean(projectId),
    retry: false,
  });
}
