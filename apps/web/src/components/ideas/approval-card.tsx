import { ShieldCheck } from "@solar-icons/react";
import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { formatArgs, type RunApproval } from "@/lib/ideas";

type ApprovalCardProps = {
  approval: RunApproval;
  busy: boolean;
  onApprove: (approval: RunApproval) => void;
  onDecline: (approval: RunApproval) => void;
};

export function ApprovalCard({
  approval,
  busy,
  onApprove,
  onDecline,
}: ApprovalCardProps) {
  const { t } = useTranslation("ideas");
  const args = formatArgs(approval.args);

  return (
    <section
      aria-label={`${t("approval.heading")}: ${approval.tool}`}
      aria-busy={busy}
      className="flex flex-col gap-3 rounded-xl bg-muted p-4"
    >
      <div className="flex items-start gap-3">
        <ShieldCheck className="mt-0.5 size-5 shrink-0 text-warning-foreground" />
        <div className="flex min-w-0 flex-col gap-1.5">
          <p className="text-xs font-medium text-muted-foreground">
            {t("approval.heading")}
          </p>
          <Badge variant="warning" className="self-start">
            {approval.tool}
          </Badge>
        </div>
      </div>
      {args && (
        <details className="group">
          <summary className="cursor-pointer text-xs font-medium text-muted-foreground">
            {t("approval.args")}
          </summary>
          <pre className="mt-2 max-h-60 overflow-auto whitespace-pre-wrap break-words rounded-lg bg-card p-3 text-xs">
            {args}
          </pre>
        </details>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <Button
          className="border-transparent"
          size="lg"
          disabled={busy}
          onClick={() => onApprove(approval)}
        >
          {t("approval.approve")}
        </Button>
        <Button
          size="lg"
          variant="secondary"
          disabled={busy}
          onClick={() => onDecline(approval)}
        >
          {t("approval.decline")}
        </Button>
        {busy && (
          <span className="flex items-center gap-2 text-xs text-muted-foreground">
            <Spinner size={14} />
            {t("approval.busy")}
          </span>
        )}
      </div>
    </section>
  );
}
