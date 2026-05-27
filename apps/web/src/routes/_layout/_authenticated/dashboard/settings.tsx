import { AltArrowLeft as ChevronLeft } from "@solar-icons/react";
import {
  createFileRoute,
  Outlet,
  useLocation,
  useNavigate,
} from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import useGetProjects from "@/hooks/queries/project/use-get-projects";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { cn } from "@/lib/cn";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings",
)({
  component: SettingsLayout,
});

function SettingsLayout() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { data: workspace } = useActiveWorkspace();
  const { data: projects } = useGetProjects({
    workspaceId: workspace?.id ?? "",
  });

  const getActiveTab = () => {
    const pathname = location.pathname;
    if (pathname.includes("/dashboard/settings/account")) {
      return "account";
    }
    if (pathname.includes("/dashboard/settings/workspace")) {
      return "workspace";
    }
    if (pathname.includes("/dashboard/settings/projects")) {
      return "project";
    }
    return "account";
  };

  const activeTab = getActiveTab();

  const tabs = [
    {
      id: "account",
      label: t("settings:account"),
      to: "/dashboard/settings/account/information" as const,
      disabled: false,
    },
    {
      id: "workspace",
      label: t("navigation:page.settingsWorkspaceTab"),
      to: "/dashboard/settings/workspace/general" as const,
      disabled: false,
    },
    {
      id: "project",
      label: t("navigation:sidebar.projects"),
      to: "/dashboard/settings/projects" as const,
      disabled: projects?.length === 0,
    },
  ];

  return (
    <>
      <PageTitle title={t("navigation:page.settingsTitle")} />
      <div className="flex flex-col w-full h-full overflow-hidden">
        <div className="px-4 py-4 md:px-8 md:py-5">
          <Button
            variant="ghost"
            size="sm"
            className="mb-3 -ml-2 text-muted-foreground"
            onClick={() =>
              navigate({
                to: "/dashboard/workspace/$workspaceId",
                params: { workspaceId: workspace?.id ?? "" },
              })
            }
          >
            <ChevronLeft className="size-4" />
            {t("navigation:page.backToWorkspace")}
          </Button>

          <h1 className="text-lg font-semibold">
            {t("navigation:page.settingsTitle")}
          </h1>

          <div className="mt-4 flex items-center gap-1.5 border-b border-border pb-3">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                type="button"
                disabled={tab.disabled}
                onClick={() => navigate({ to: tab.to })}
                className={cn(
                  "inline-flex h-9 items-center rounded-lg px-3.5 text-sm font-medium transition-colors",
                  activeTab === tab.id
                    ? "bg-foreground text-background"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                  tab.disabled && "opacity-40 pointer-events-none",
                )}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-4 py-6 md:px-8">
          <Outlet />
        </div>
      </div>
    </>
  );
}
