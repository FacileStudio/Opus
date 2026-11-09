export type PullRequest = {
  id: number;
  taskId: number;
  repoFullName: string;
  prNumber: number;
  title: string;
  url: string;
  state: "open" | "draft" | "merged" | "closed";
  headBranch: string;
  baseBranch: string;
  authorGithubId?: string;
  createdAt: string;
  updatedAt: string;
};
