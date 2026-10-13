import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { NotificationPreferencesSettings } from "@/components/account/notification-preferences-settings";
import PageTitle from "@/components/page-title";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/account/notifications",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();

  return (
    <>
      <PageTitle title={t("settings:notificationsPage.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">
            {t("settings:notificationsPage.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("settings:notificationsPage.subtitle")}
          </p>
        </div>

        <NotificationPreferencesSettings />
      </div>
    </>
  );
}
