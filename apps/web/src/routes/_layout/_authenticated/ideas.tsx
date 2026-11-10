import { Refresh } from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import Layout from "@/components/common/layout";
import { IdeaCard } from "@/components/ideas/idea-card";
import { IdeasIcon } from "@/components/ideas/ideas-icon";
import { IdeasUpsell } from "@/components/ideas/ideas-upsell";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Spinner } from "@/components/ui/spinner";
import {
  ideasPath,
  ideasQueryKey,
  useIdeaRun,
  useIdeas,
} from "@/hooks/queries/ideas/use-ideas";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { apiRequest } from "@/lib/auth-client";
import {
  type Idea,
  type IdeaRun,
  ideasView,
  loadRunIds,
  type RunApproval,
  relativeTime,
  saveRunIds,
} from "@/lib/ideas";

export const Route = createFileRoute("/_layout/_authenticated/ideas")({
  component: IdeasPageRoute,
});

type CardProps = {
  idea: Idea;
  runId: string | undefined;
  onRun: (idea: Idea) => Promise<string | null>;
  onDismiss: (idea: Idea) => Promise<void>;
};

function IdeaCardContainer({ idea, runId, onRun, onDismiss }: CardProps) {
  const queryClient = useQueryClient();
  const { data: run, error: runLoadError } = useIdeaRun(runId);
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const guard = async (action: () => Promise<string | null | undefined>) => {
    setBusy(true);
    setActionError(null);
    try {
      const failure = await action();
      if (failure) setActionError(failure);
    } finally {
      setBusy(false);
    }
  };

  const decide = (approval: RunApproval, approve: boolean) =>
    guard(async () => {
      if (!runId) return null;
      const result = await apiRequest<void>(
        `/ideas/runs/${runId}/approvals/${encodeURIComponent(approval.call_id)}`,
        { method: "POST", body: JSON.stringify({ approve }) },
        { mutation: true },
      );
      if (result.error) return result.error.message;
      await queryClient.invalidateQueries({
        queryKey: ["ideas", "run", runId],
      });
      return null;
    });

  return (
    <IdeaCard
      idea={idea}
      run={run as IdeaRun | undefined}
      runStarted={!!runId && !run && !runLoadError}
      runError={actionError}
      busy={busy}
      onRun={() => guard(() => onRun(idea))}
      onDismiss={() =>
        guard(async () => {
          await onDismiss(idea);
          return null;
        })
      }
      onApprove={(_, approval) => decide(approval, true)}
      onDecline={(_, approval) => decide(approval, false)}
    />
  );
}

function IdeasPageRoute() {
  const { t, i18n } = useTranslation("ideas");
  const queryClient = useQueryClient();
  const { data: workspace } = useActiveWorkspace();
  const workspaceId = workspace?.id;
  const { data, isLoading, error, refetch } = useIdeas(
    workspaceId,
    !!workspace,
  );
  const [runIds, setRunIds] = useState<Record<string, string>>(loadRunIds);
  const [generateError, setGenerateError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);
  const key = ideasQueryKey(workspaceId);

  const generating = starting || !!data?.generating;
  const errorMessage = generateError ?? (error ? error.message : null);
  const ideas = data?.ideas ?? [];
  const view = ideasView({
    count: ideas.length,
    loading: isLoading || !workspace,
    generating,
    error: errorMessage,
  });

  const generate = async () => {
    setStarting(true);
    setGenerateError(null);
    try {
      const result = await apiRequest<{ batch_id: string }>(
        ideasPath("/generate", workspaceId),
        { method: "POST" },
        { mutation: true },
      );
      if (result.error && result.error.status !== 409) {
        setGenerateError(result.error.message);
        return;
      }
      await queryClient.invalidateQueries({ queryKey: key });
    } finally {
      setStarting(false);
    }
  };

  const runIdea = async (idea: Idea): Promise<string | null> => {
    const result = await apiRequest<{ run_id: string }>(
      ideasPath(`/${encodeURIComponent(idea.id)}/run`, workspaceId),
      { method: "POST" },
      { mutation: true },
    );
    if (result.error) return result.error.message;
    const next = { ...runIds, [idea.id]: result.data.run_id };
    setRunIds(next);
    saveRunIds(next);
    return null;
  };

  const dismissIdea = async (idea: Idea) => {
    const result = await apiRequest<void>(
      ideasPath(`/${encodeURIComponent(idea.id)}/dismiss`, workspaceId),
      { method: "POST" },
      { mutation: true },
    );
    if (result.error) {
      setGenerateError(result.error.message);
      return;
    }
    await queryClient.invalidateQueries({ queryKey: key });
  };

  const last = data?.last_batch_at
    ? t("lastGenerated", {
        time: relativeTime(data.last_batch_at, i18n.language),
      })
    : t("neverGenerated");
  const free = !!data && !data.ai_access;

  return (
    <>
      <PageTitle title={t("title")} />
      <Layout>
        <Layout.Header>
          <div className="flex w-full items-center gap-1">
            <SidebarTrigger className="-ml-1 h-6 w-6" />
            <Separator
              orientation="vertical"
              className="mx-1.5 data-[orientation=vertical]:h-2.5"
            />
            <h1 className="text-xs text-card-foreground">{t("title")}</h1>
          </div>
        </Layout.Header>
        <Layout.Content>
          <div className="flex flex-col gap-6 p-4 sm:p-6">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="flex min-w-0 flex-col gap-1">
                <h2 className="m-0 text-xl font-semibold">{t("title")}</h2>
                <p className="text-sm text-muted-foreground">
                  {t("description")}
                </p>
                {!free && (
                  <p className="text-xs text-muted-foreground">{last}</p>
                )}
              </div>
              {!free && (
                <Button
                  className="border-transparent"
                  size="lg"
                  disabled={generating}
                  aria-busy={generating}
                  onClick={generate}
                  data-testid="ideas-generate"
                >
                  {generating ? <Spinner size={16} /> : <IdeasIcon />}
                  {generating ? t("generating") : t("generate")}
                </Button>
              )}
            </div>

            <p className="sr-only" role="status">
              {generating ? t("generating") : ""}
            </p>

            {free ? (
              <IdeasUpsell upgradeUrl={data?.upgrade_url} />
            ) : view === "loading" ? (
              <div className="flex min-h-[30vh] items-center justify-center">
                <Spinner />
              </div>
            ) : view === "error" ? (
              <Placeholder
                title={t("error.title")}
                description={errorMessage ?? undefined}
              >
                <Button
                  className="border-transparent"
                  size="lg"
                  onClick={() => {
                    setGenerateError(null);
                    void refetch();
                  }}
                >
                  <Refresh />
                  {t("error.retry")}
                </Button>
              </Placeholder>
            ) : view === "busy" ? (
              <Placeholder
                title={t("busy.title")}
                description={t("busy.description")}
              >
                <Spinner />
              </Placeholder>
            ) : view === "empty" ? (
              <Placeholder
                title={t("empty.title")}
                description={t("empty.description")}
              >
                <Button
                  className="border-transparent"
                  size="lg"
                  onClick={generate}
                >
                  <IdeasIcon />
                  {t("generate")}
                </Button>
              </Placeholder>
            ) : (
              <>
                {errorMessage && (
                  <p
                    role="alert"
                    className="rounded-xl bg-destructive/8 p-3 text-sm text-destructive-foreground"
                  >
                    {errorMessage}
                  </p>
                )}
                <ul
                  className="grid gap-4 md:grid-cols-2 xl:grid-cols-3"
                  aria-label={t("list")}
                >
                  {ideas.map((idea) => (
                    <li key={idea.id} className="min-w-0">
                      <IdeaCardContainer
                        idea={idea}
                        runId={runIds[idea.id]}
                        onRun={runIdea}
                        onDismiss={dismissIdea}
                      />
                    </li>
                  ))}
                </ul>
              </>
            )}
          </div>
        </Layout.Content>
      </Layout>
    </>
  );
}

function Placeholder({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
      <span className="flex size-14 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <IdeasIcon size={28} />
      </span>
      <h3 className="m-0 text-base font-semibold">{title}</h3>
      {description && (
        <p className="max-w-md text-sm text-muted-foreground">{description}</p>
      )}
      {children}
    </div>
  );
}
