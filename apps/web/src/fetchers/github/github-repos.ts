import { getApiUrl } from "@/fetchers/get-api-url";
import type {
  GitHubIssueState,
  GitHubIssues,
  GitHubRepo,
} from "@/types/github";

async function githubRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const method = (init?.method ?? "GET").toUpperCase();
  const response = await fetch(getApiUrl(path), {
    ...init,
    credentials: "include",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(method === "GET" ? {} : { "X-Facile-CSRF": "1" }),
    },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.error?.message ?? response.statusText);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json();
}

export function getGitHubRepos(projectId: string) {
  return githubRequest<GitHubRepo[]>(`/project/${projectId}/github-repos`);
}

export function linkGitHubRepo(projectId: string, repo: string) {
  return githubRequest<GitHubRepo>(`/project/${projectId}/github-repos`, {
    method: "POST",
    body: JSON.stringify({ repo }),
  });
}

export function unlinkGitHubRepo(projectId: string, repoId: number) {
  return githubRequest<void>(`/project/${projectId}/github-repos/${repoId}`, {
    method: "DELETE",
  });
}

export function getGitHubIssues(projectId: string, state: GitHubIssueState) {
  return githubRequest<GitHubIssues>(
    `/project/${projectId}/github-issues?state=${state}`,
  );
}
