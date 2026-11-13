import { ChatRound } from "@solar-icons/react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import ProjectLayout from "@/components/common/project-layout";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import {
  useGetGitHubIssues,
  useGetGitHubRepos,
} from "@/hooks/queries/github/use-github-repos";
import useGetProject from "@/hooks/queries/project/use-get-project";
import { cn } from "@/lib/cn";
import type { GitHubIssue, GitHubIssueState } from "@/types/github";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/project/$projectId/issues",
)({
  component: RouteComponent,
});

const states: GitHubIssueState[] = ["open", "closed"];

function RouteComponent() {
  const { t } = useTranslation();
  const { projectId, workspaceId } = Route.useParams();
  const [state, setState] = useState<GitHubIssueState>("open");
  const { data: project } = useGetProject({ id: projectId, workspaceId });
  const { data: repos } = useGetGitHubRepos(projectId);
  const { data, error, isLoading } = useGetGitHubIssues(projectId, state);

  return (
    <ProjectLayout
      projectId={projectId}
      workspaceId={workspaceId}
      activeView="issues"
    >
      <PageTitle
        title={t("tasks:issues.pageTitle", { name: project?.name })}
        hideAppName
      />
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-3 p-2">
        <div className="flex items-center justify-between gap-2">
          <h1 className="text-sm font-semibold">{t("tasks:issues.title")}</h1>
          <div className="inline-flex gap-0.5 rounded-full p-0.5">
            {states.map((value) => (
              <Button
                key={value}
                size="xs"
                variant={state === value ? "secondary" : "ghost"}
                className={cn(
                  "h-6 rounded-full px-2.5 text-xs",
                  state !== value && "text-muted-foreground",
                )}
                onClick={() => setState(value)}
              >
                {t(`tasks:issues.${value}`)}
              </Button>
            ))}
          </div>
        </div>

        {error && (
          <p className="rounded-xl bg-card p-4 text-sm text-muted-foreground">
            {error.message}
          </p>
        )}

        {repos && repos.length === 0 && (
          <div className="rounded-xl bg-card p-6 text-center space-y-2">
            <p className="text-sm font-medium">{t("tasks:issues.noRepos")}</p>
            <p className="text-sm text-muted-foreground">
              {t("tasks:issues.noReposSubtitle")}
            </p>
            <Button
              size="sm"
              variant="secondary"
              render={
                <Link
                  to="/dashboard/settings/projects/$projectId/integrations"
                  params={{ projectId }}
                />
              }
            >
              {t("tasks:issues.openSettings")}
            </Button>
          </div>
        )}

        {data?.unavailable.map((repo) => (
          <p
            key={repo.repo}
            className="rounded-xl bg-card px-4 py-2 text-xs text-muted-foreground"
          >
            {repo.reason}
          </p>
        ))}

        {!isLoading &&
          !error &&
          repos &&
          repos.length > 0 &&
          data?.issues.length === 0 && (
            <p className="rounded-xl bg-card p-6 text-center text-sm text-muted-foreground">
              {t("tasks:issues.empty")}
            </p>
          )}

        <ul className="flex flex-col gap-1">
          {data?.issues.map((issue) => (
            <IssueRow key={`${issue.repo}#${issue.number}`} issue={issue} />
          ))}
        </ul>
      </div>
    </ProjectLayout>
  );
}

function labelStyle(color: string) {
  return /^[0-9a-f]{6}$/i.test(color)
    ? { backgroundColor: `#${color}33` }
    : undefined;
}

function IssueRow({ issue }: { issue: GitHubIssue }) {
  const { t } = useTranslation();
  return (
    <li className="rounded-xl bg-card px-4 py-3">
      <a
        href={issue.url}
        target="_blank"
        rel="noreferrer"
        className="flex items-start justify-between gap-3"
      >
        <div className="min-w-0 space-y-1">
          <p className="truncate text-sm font-medium hover:underline">
            {issue.title}
          </p>
          <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
            <span>
              {t("tasks:issues.openedBy", {
                repo: issue.repo,
                number: issue.number,
                author: issue.author?.login ?? "ghost",
              })}
            </span>
            {issue.labels.map((label) => (
              <span
                key={label.name}
                className="rounded-full px-2 py-0.5 text-[10px] font-medium"
                style={labelStyle(label.color)}
              >
                {label.name}
              </span>
            ))}
          </div>
        </div>
        {issue.comments > 0 && (
          <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
            <ChatRound size={14} />
            {issue.comments}
          </span>
        )}
      </a>
    </li>
  );
}
