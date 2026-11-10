import {
  Bolt,
  CalendarMark,
  CheckCircle,
  CloseCircle,
  History,
  Letter,
  Magnifer,
  Refresh,
  Widget,
} from "@solar-icons/react";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { ApprovalCard } from "@/components/ideas/approval-card";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import {
  type Idea,
  type IdeaRun,
  ideaKind,
  isSafeHref,
  type RunApproval,
  type RunState,
} from "@/lib/ideas";

const KIND_ICONS = {
  mail: Letter,
  task: CheckCircle,
  event: CalendarMark,
  follow_up: History,
  research: Magnifer,
  other: Widget,
} as const;

const STATE_BADGES = {
  running: "info",
  needs_approval: "warning",
  done: "success",
  failed: "error",
} as const;

type IdeaCardProps = {
  idea: Idea;
  run: IdeaRun | undefined;
  runStarted: boolean;
  runError: string | null;
  busy: boolean;
  onRun: (idea: Idea) => void;
  onDismiss: (idea: Idea) => void;
  onApprove: (idea: Idea, approval: RunApproval) => void;
  onDecline: (idea: Idea, approval: RunApproval) => void;
};

function linkButton(
  href: string,
  label: ReactNode,
  variant: "secondary" | "ghost",
) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className={buttonVariants({ size: "lg", variant })}
    >
      {label}
    </a>
  );
}

export function IdeaCard({
  idea,
  run,
  runStarted,
  runError,
  busy,
  onRun,
  onDismiss,
  onApprove,
  onDecline,
}: IdeaCardProps) {
  const { t } = useTranslation("ideas");
  const kind = ideaKind(idea.kind);
  const KindIcon = KIND_ICONS[kind];
  const state: RunState | "idle" =
    run?.state ?? (runStarted ? "running" : "idle");
  const approval = run?.pending_approvals?.[0];
  const outputs = (run?.outputs ?? []).filter((output) =>
    isSafeHref(output.url),
  );
  const sources = (idea.sources ?? []).filter((source) =>
    isSafeHref(source.ref),
  );
  const continueHref = isSafeHref(run?.continue_url) ? run.continue_url : null;
  const badge = state === "idle" ? null : STATE_BADGES[state];

  return (
    <Card className="h-full gap-4 p-5" data-testid="idea-card">
      <div className="flex items-start gap-3">
        <span
          className="flex size-9 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground"
          aria-hidden="true"
        >
          <KindIcon className="size-4.5" />
        </span>
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge variant="secondary">{t(`kind.${kind}`)}</Badge>
            {badge && <Badge variant={badge}>{t(`state.${state}`)}</Badge>}
          </div>
          <h2 className="m-0 break-words text-base font-medium">
            {idea.title}
          </h2>
          <p className="break-words text-sm text-muted-foreground">
            {idea.why}
          </p>
        </div>
      </div>

      {sources.length > 0 && (
        <ul className="flex flex-wrap gap-2" aria-label={t("card.sources")}>
          {sources.map((source) => (
            <li key={source.ref} className="max-w-full">
              <a
                href={source.ref}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-block max-w-full truncate rounded-full bg-muted px-3 py-1 text-xs text-muted-foreground hover:text-foreground"
              >
                {source.label}
              </a>
            </li>
          ))}
        </ul>
      )}

      {idea.draft && (
        <details>
          <summary className="cursor-pointer text-xs font-medium text-muted-foreground">
            {t("card.draft")}
          </summary>
          <p className="mt-2 whitespace-pre-wrap break-words rounded-lg bg-muted p-3 text-sm">
            {idea.draft}
          </p>
        </details>
      )}

      <div className="flex flex-col gap-3" aria-live="polite">
        {state === "running" && (
          <p className="flex items-center gap-2 text-sm text-muted-foreground">
            <Spinner size={16} />
            {t("state.running")}
          </p>
        )}
        {state === "needs_approval" && approval && (
          <ApprovalCard
            approval={approval}
            busy={busy}
            onApprove={(value) => onApprove(idea, value)}
            onDecline={(value) => onDecline(idea, value)}
          />
        )}
        {state === "done" && (
          <>
            {run?.summary && (
              <p className="break-words text-sm">{run.summary}</p>
            )}
            {outputs.length > 0 && (
              <ul
                className="flex flex-wrap gap-2"
                aria-label={t("card.outputs")}
                data-testid="idea-outputs"
              >
                {outputs.map((output) => (
                  <li key={output.call_id} className="max-w-full">
                    {linkButton(
                      output.url,
                      output.preview || output.tool,
                      "secondary",
                    )}
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
        {state === "failed" && (
          <p className="rounded-lg bg-destructive/8 p-3 text-sm text-destructive-foreground">
            {runError || t("card.failed")}
          </p>
        )}
        {state === "idle" && runError && (
          <p className="rounded-lg bg-destructive/8 p-3 text-sm text-destructive-foreground">
            {runError}
          </p>
        )}
      </div>

      <div className="mt-auto flex flex-wrap items-center gap-2">
        {state === "idle" && (
          <Button
            className="border-transparent"
            size="lg"
            disabled={busy}
            onClick={() => onRun(idea)}
          >
            <Bolt />
            {t("card.run")}
          </Button>
        )}
        {state === "failed" && (
          <Button
            className="border-transparent"
            size="lg"
            disabled={busy}
            onClick={() => onRun(idea)}
          >
            <Refresh />
            {t("card.retry")}
          </Button>
        )}
        {state !== "idle" &&
          continueHref &&
          linkButton(continueHref, t("card.continue"), "ghost")}
        {state !== "running" && !(state === "needs_approval" && approval) && (
          <Button
            size="lg"
            variant="ghost"
            disabled={busy}
            onClick={() => onDismiss(idea)}
          >
            <CloseCircle />
            {t("card.dismiss")}
          </Button>
        )}
      </div>
    </Card>
  );
}
