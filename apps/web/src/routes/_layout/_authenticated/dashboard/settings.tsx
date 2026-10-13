import {
  AltArrowLeft as ChevronLeft,
  Folder2,
  Settings,
  User,
  UsersGroupTwoRounded,
} from "@solar-icons/react";
import {
  createFileRoute,
  Outlet,
  useLocation,
  useNavigate,
} from "@tanstack/react-router";
import type { ComponentType, SVGProps } from "react";
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
    if (pathname.includes("/dashboard/settings/workspace/members")) {
      return "members";
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

  const tabs: Array<{
    id: string;
    label: string;
    icon: ComponentType<SVGProps<SVGSVGElement>>;
    to: string;
    disabled: boolean;
  }> = [
    {
      id: "account",
      label: t("settings:account"),
      icon: User,
      to: "/dashboard/settings/account/information",
      disabled: false,
    },
    {
      id: "workspace",
      label: t("navigation:page.settingsWorkspaceTab"),
      icon: Settings,
      to: "/dashboard/settings/workspace/general",
      disabled: false,
    },
    {
      id: "members",
      label: t("team:members.pageTitle"),
      icon: UsersGroupTwoRounded,
      to: "/dashboard/settings/workspace/members",
      disabled: false,
    },
    {
      id: "project",
      label: t("navigation:sidebar.projects"),
      icon: Folder2,
      to: "/dashboard/settings/projects",
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

          <div className="flex flex-col gap-4">
            <h1 className="text-2xl font-semibold">
              {t("navigation:page.settingsTitle")}
            </h1>

            <div className="flex w-full items-center gap-1 overflow-x-auto">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  type="button"
                  disabled={tab.disabled}
                  onClick={() => navigate({ to: tab.to })}
                  className={cn(
                    "inline-flex min-h-11 shrink-0 items-center gap-2 rounded-full px-4 text-sm font-medium transition-colors",
                    activeTab === tab.id
                      ? "bg-foreground text-background"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground",
                    tab.disabled && "pointer-events-none opacity-50",
                  )}
                >
                  <tab.icon className="size-4" />
                  {tab.label}
                </button>
              ))}
            </div>

            <div className="border-b border-border" />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-4 py-6 md:px-8">
          <Outlet />
        </div>
      </div>
    </>
  );
}
