export type GitHubRepo = {
  id: number;
  projectId: number;
  fullName: string;
  defaultBranch: string;
  url: string;
  createdAt: string;
};

export type GitHubIssueState = "open" | "closed";

export type GitHubIssue = {
  repo: string;
  number: number;
  title: string;
  url: string;
  state: GitHubIssueState;
  comments: number;
  labels: { name: string; color: string }[];
  author: { login: string; avatarUrl: string; url: string } | null;
  createdAt: string;
  updatedAt: string;
};

export type GitHubIssues = {
  issues: GitHubIssue[];
  unavailable: { repo: string; reason: string }[];
};
