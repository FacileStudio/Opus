import {
  CheckCircle,
  ShieldWarning,
  TrashBinMinimalistic,
} from "@solar-icons/react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { z } from "zod/v4";
import { SignInWithFacile } from "@/components/auth/sign-in-with-facile";
import Layout from "@/components/common/layout";
import PageTitle from "@/components/page-title";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Spinner } from "@/components/ui/spinner";
import { ApiRequestError, apiRequest } from "@/lib/auth-client";
import { relativeTime } from "@/lib/ideas";
import { clientName, type PendingAction } from "@/lib/pending";

const pendingSearchSchema = z.object({
  action: z.string().optional(),
});

export const Route = createFileRoute("/_layout/_authenticated/pending")({
  component: PendingPage,
  validateSearch: pendingSearchSchema,
});

const QUERY_KEY = ["mcp", "pending"] as const;

const NOT_OWN_SESSION = 403;

function usePending() {
  return useQuery({
    queryKey: QUERY_KEY,
    queryFn: async () => {
      const result = await apiRequest<{ actions: PendingAction[] }>(
        "/mcp/pending",
      );
      if (result.error) throw new ApiRequestError(result.error);
      return result.data.actions ?? [];
    },
  });
}

function PendingPage() {
  const { t, i18n } = useTranslation("pending");
  const { action: highlighted } = Route.useSearch();
  const queryClient = useQueryClient();
  const { data: actions, isLoading, error, refetch } = usePending();
  const [busy, setBusy] = useState<string | null>(null);
  const [notice, setNotice] = useState<{
    tone: "ok" | "error";
    text: string;
    reauth?: boolean;
  } | null>(null);
  const reauth =
    error instanceof ApiRequestError && error.status === NOT_OWN_SESSION;

  const answer = async (item: PendingAction, verb: "confirm" | "dismiss") => {
    setBusy(item.id);
    setNotice(null);
    const result = await apiRequest<{ status: string }>(
      `/mcp/pending/${encodeURIComponent(item.id)}/${verb}`,
      { method: "POST" },
      { mutation: true },
    );
    setBusy(null);
    if (result.error?.status === NOT_OWN_SESSION) {
      setNotice({ tone: "error", text: t("reauth.description"), reauth: true });
    } else if (result.error) {
      setNotice({ tone: "error", text: result.error.message });
    } else {
      setNotice({
        tone: "ok",
        text: t(verb === "confirm" ? "done.deleted" : "done.dismissed"),
      });
    }
    await queryClient.invalidateQueries({ queryKey: QUERY_KEY });
  };

  const list = actions ?? [];
  const ordered = highlighted
    ? [
        ...list.filter((a) => a.id === highlighted),
        ...list.filter((a) => a.id !== highlighted),
      ]
    : list;

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
            <div className="flex min-w-0 flex-col gap-1">
              <h2 className="m-0 text-xl font-semibold">{t("title")}</h2>
              <p className="max-w-prose text-sm text-muted-foreground">
                {t("description")}
              </p>
            </div>

            {notice && (
              <div
                role={notice.tone === "error" ? "alert" : "status"}
                className={
                  notice.tone === "error"
                    ? "flex flex-col items-start gap-3 rounded-xl bg-destructive/8 p-3 text-sm text-destructive-foreground"
                    : "rounded-xl bg-muted p-3 text-sm"
                }
              >
                <p className="m-0">{notice.text}</p>
                {notice.reauth && <SignInWithFacile />}
              </div>
            )}

            {isLoading ? (
              <div className="flex min-h-[30vh] items-center justify-center">
                <Spinner />
              </div>
            ) : reauth ? (
              <Empty
                title={t("reauth.title")}
                description={t("reauth.description")}
              >
                <SignInWithFacile size="lg" />
              </Empty>
            ) : error ? (
              <Empty title={t("error.title")} description={error.message}>
                <Button
                  className="border-transparent"
                  size="lg"
                  onClick={() => void refetch()}
                >
                  {t("error.retry")}
                </Button>
              </Empty>
            ) : ordered.length === 0 ? (
              <Empty
                title={t("empty.title")}
                description={t("empty.description")}
              />
            ) : (
              <ul className="flex flex-col gap-4" aria-label={t("list")}>
                {ordered.map((item) => (
                  <li key={item.id} className="min-w-0">
                    <PendingCard
                      item={item}
                      highlighted={item.id === highlighted}
                      busy={busy === item.id}
                      locale={i18n.language}
                      onConfirm={() => void answer(item, "confirm")}
                      onDismiss={() => void answer(item, "dismiss")}
                    />
                  </li>
                ))}
              </ul>
            )}
          </div>
        </Layout.Content>
      </Layout>
    </>
  );
}

type CardProps = {
  item: PendingAction;
  highlighted: boolean;
  busy: boolean;
  locale: string;
  onConfirm: () => void;
  onDismiss: () => void;
};

function PendingCard({
  item,
  highlighted,
  busy,
  locale,
  onConfirm,
  onDismiss,
}: CardProps) {
  const { t } = useTranslation("pending");
  const target = item.target;
  const ready = item.state === "ready";

  return (
    <Card
      className={`gap-4 p-5 ${highlighted ? "bg-muted" : ""}`}
      data-testid="pending-action"
      aria-current={highlighted ? "true" : undefined}
    >
      <div className="flex flex-wrap items-center gap-2">
        <Badge
          variant={ready ? "secondary" : "warning"}
          className="whitespace-nowrap"
        >
          {t(`state.${item.state}`)}
        </Badge>
        <span className="text-xs text-muted-foreground">
          {t("requestedBy", { client: clientName(item.client_id) })} ·{" "}
          {relativeTime(item.created_at, locale)}
        </span>
      </div>
      <p className="m-0 break-words text-base font-medium">{item.summary}</p>
      {target && (
        <a
          href={target.url}
          className="flex min-w-0 flex-wrap items-center gap-2 text-sm text-muted-foreground hover:underline"
        >
          {target.ref && (
            <Badge variant="secondary" className="whitespace-nowrap">
              {target.ref}
            </Badge>
          )}
          <span className="break-words">{target.title}</span>
          {target.project_name && <span>· {target.project_name}</span>}
          {target.type === "project" && (
            <span>· {t("tasks", { count: target.task_count ?? 0 })}</span>
          )}
        </a>
      )}
      {!ready && (
        <p className="m-0 flex items-start gap-2 text-sm text-muted-foreground">
          <ShieldWarning className="mt-0.5 shrink-0" />
          {t(`stateHelp.${item.state}`)}
        </p>
      )}
      <div className="flex flex-wrap gap-2">
        <Button
          variant="destructive"
          className="border-transparent"
          size="lg"
          disabled={!ready || busy}
          aria-busy={busy}
          onClick={onConfirm}
          data-testid="pending-confirm"
        >
          {busy ? <Spinner size={16} /> : <TrashBinMinimalistic />}
          {t("confirm")}
        </Button>
        <Button
          variant="ghost"
          className="border-transparent"
          size="lg"
          disabled={busy}
          onClick={onDismiss}
          data-testid="pending-dismiss"
        >
          {t("dismiss")}
        </Button>
      </div>
    </Card>
  );
}

function Empty({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
      <span className="flex size-14 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <CheckCircle size={28} />
      </span>
      <h3 className="m-0 text-base font-semibold">{title}</h3>
      {description && (
        <p className="max-w-md text-sm text-muted-foreground">{description}</p>
      )}
      {children}
    </div>
  );
}
