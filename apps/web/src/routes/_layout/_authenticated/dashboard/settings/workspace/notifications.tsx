import { Radio } from "@solar-icons/react";
import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace/notifications",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();

  return (
    <>
      <PageTitle title={t("settings:workspaceNotifications.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">
            {t("settings:workspaceNotifications.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("settings:workspaceNotifications.subtitle")}
          </p>
        </div>

        <div className="rounded-xl border border-border bg-background p-6 space-y-2">
          <div className="flex items-center gap-2 font-medium text-sm">
            <Radio className="size-4 text-muted-foreground" />
            <span>Event Bus</span>
          </div>
          <p className="text-sm text-muted-foreground">
            Workspace event notifications and webhooks are dispatched via
            Antenne.
          </p>
        </div>
      </div>
    </>
  );
}
