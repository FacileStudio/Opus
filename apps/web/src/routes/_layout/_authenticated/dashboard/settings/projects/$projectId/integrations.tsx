import { Radio, TrashBinMinimalistic } from "@solar-icons/react";
import { createFileRoute, useParams } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  useLinkGitHubRepo,
  useUnlinkGitHubRepo,
} from "@/hooks/mutations/github/use-github-repo-mutations";
import { useGetGitHubRepos } from "@/hooks/queries/github/use-github-repos";
import { useWorkspacePermission } from "@/hooks/use-workspace-permission";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/projects/$projectId/integrations",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();

  return (
    <>
      <PageTitle title={t("settings:projectIntegrations.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">
            {t("settings:projectIntegrations.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("settings:projectIntegrations.subtitle")}
          </p>
        </div>

        <GitHubSection />

        <div className="rounded-xl bg-background p-6 space-y-2">
          <div className="flex items-center gap-2 font-medium text-sm">
            <Radio className="size-4 text-muted-foreground" />
            <span>Event Bus</span>
          </div>
          <p className="text-sm text-muted-foreground">
            External integrations and webhook events are routed via Antenne.
          </p>
        </div>
      </div>
    </>
  );
}

function GitHubSection() {
  const { t } = useTranslation();
  const { projectId = "" } = useParams({ strict: false });
  const { isAdmin } = useWorkspacePermission();
  const { data: repos = [], error, isLoading } = useGetGitHubRepos(projectId);
  const linkRepo = useLinkGitHubRepo(projectId);
  const unlinkRepo = useUnlinkGitHubRepo(projectId);
  const [name, setName] = useState("");

  const handleLink = async (event: FormEvent) => {
    event.preventDefault();
    const repo = name.trim();
    if (!repo) return;
    try {
      await linkRepo.mutateAsync(repo);
      setName("");
      toast.success(t("settings:projectIntegrations.githubLinked"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e));
    }
  };

  const handleUnlink = async (repoId: number) => {
    try {
      await unlinkRepo.mutateAsync(repoId);
      toast.success(t("settings:projectIntegrations.githubUnlinked"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div className="rounded-xl bg-background p-6 space-y-4">
      <div className="space-y-1">
        <h2 className="font-medium text-sm">
          {t("settings:projectIntegrations.githubSectionTitle")}
        </h2>
        <p className="text-sm text-muted-foreground">
          {t("settings:projectIntegrations.githubSectionSubtitle")}
        </p>
      </div>

      {error ? (
        <p className="text-sm text-muted-foreground">{error.message}</p>
      ) : (
        <>
          {!isLoading && repos.length === 0 && (
            <p className="text-sm text-muted-foreground">
              {t("settings:projectIntegrations.githubNoRepos")}
            </p>
          )}
          <ul className="space-y-2">
            {repos.map((repo) => (
              <li
                key={repo.id}
                className="flex items-center justify-between gap-2 rounded-lg bg-muted/50 px-3 py-2"
              >
                <div className="min-w-0">
                  <a
                    href={repo.url}
                    target="_blank"
                    rel="noreferrer"
                    className="block truncate text-sm font-medium hover:underline"
                  >
                    {repo.fullName}
                  </a>
                  <p className="text-xs text-muted-foreground">
                    {t("settings:projectIntegrations.githubDefaultBranch", {
                      branch: repo.defaultBranch,
                    })}
                  </p>
                </div>
                {isAdmin && (
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={unlinkRepo.isPending}
                    onClick={() => handleUnlink(repo.id)}
                  >
                    <TrashBinMinimalistic className="size-4" />
                    {t("settings:projectIntegrations.githubUnlink")}
                  </Button>
                )}
              </li>
            ))}
          </ul>
          {isAdmin ? (
            <form onSubmit={handleLink} className="space-y-2">
              <div className="flex gap-2">
                <Input
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder={t(
                    "settings:projectIntegrations.githubRepoPlaceholder",
                  )}
                  aria-label={t(
                    "settings:projectIntegrations.githubRepoPlaceholder",
                  )}
                />
                <Button
                  type="submit"
                  disabled={linkRepo.isPending || !name.trim()}
                >
                  {t("settings:projectIntegrations.githubLink")}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                {t("settings:projectIntegrations.githubLinkHint")}
              </p>
            </form>
          ) : (
            <p className="text-xs text-muted-foreground">
              {t("settings:projectIntegrations.githubAdminOnly")}
            </p>
          )}
        </>
      )}
    </div>
  );
}
