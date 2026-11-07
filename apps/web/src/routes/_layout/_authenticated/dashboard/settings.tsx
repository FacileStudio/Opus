import {
  Bell,
  BellBing,
  AltArrowLeft as ChevronLeft,
  AltArrowRight as ChevronRight,
  Folder2,
  Key,
  Settings,
  User,
  UsersGroupTwoRounded,
} from "@solar-icons/react";
import {
  createFileRoute,
  Link,
  Outlet,
  useLocation,
  useNavigate,
} from "@tanstack/react-router";
import type { ComponentType, SVGProps } from "react";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { NookIcon } from "@/components/icons/nook-icon";
import { SpaceIcon } from "@/components/icons/space-icon";
import PageTitle from "@/components/page-title";
import useGetProjects from "@/hooks/queries/project/use-get-projects";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { cn } from "@/lib/cn";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings",
)({
  component: SettingsLayout,
});

const BASE = "/dashboard/settings";

type Section = {
  id: string;
  label: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  to: string;
  match: string;
  hidden?: boolean;
};

const row =
  "flex min-h-11 items-center gap-2.5 rounded-fc-md px-3 text-fc-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring";

function SettingsLayout() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const { data: workspace } = useActiveWorkspace();
  const { data: projects } = useGetProjects({
    workspaceId: workspace?.id ?? "",
  });

  const sections: Section[] = [
    {
      id: "profile",
      label: t("settings:sectionProfile"),
      icon: User,
      to: `${BASE}/account/information`,
      match: `${BASE}/account/information`,
    },
    {
      id: "preferences",
      label: t("settings:preferences"),
      icon: Settings,
      to: `${BASE}/account/preferences`,
      match: `${BASE}/account/preferences`,
    },
    {
      id: "notifications",
      label: t("settings:notifications"),
      icon: Bell,
      to: `${BASE}/account/notifications`,
      match: `${BASE}/account/notifications`,
    },
    {
      id: "workspace",
      label: t("navigation:page.settingsWorkspaceTab"),
      icon: SpaceIcon,
      to: `${BASE}/workspace/general`,
      match: `${BASE}/workspace/general`,
    },
    {
      id: "members",
      label: t("team:members.pageTitle"),
      icon: UsersGroupTwoRounded,
      to: `${BASE}/workspace/members`,
      match: `${BASE}/workspace/members`,
    },
    {
      id: "workspace-notifications",
      label: t("settings:workspaceNotifications.title"),
      icon: BellBing,
      to: `${BASE}/workspace/notifications`,
      match: `${BASE}/workspace/notifications`,
    },
    {
      id: "pool",
      label: "Nook Pool",
      icon: NookIcon,
      to: `${BASE}/workspace/sync`,
      match: `${BASE}/workspace/sync`,
    },
    {
      id: "projects",
      label: t("navigation:sidebar.projects"),
      icon: Folder2,
      to: `${BASE}/projects`,
      match: `${BASE}/projects`,
      hidden: projects?.length === 0,
    },
    {
      id: "tokens",
      label: t("settings:apiKeys"),
      icon: Key,
      to: `${BASE}/account/developer`,
      match: `${BASE}/account/developer`,
    },
  ];

  const bare = pathname.replace(/\/$/, "") === BASE;
  const current = sections.find((s) => pathname.startsWith(s.match));
  const drilled = !bare;

  useEffect(() => {
    if (bare && window.matchMedia("(min-width: 64rem)").matches) {
      navigate({ to: sections[0].to, replace: true });
    }
  }, [bare, navigate, sections[0].to]);

  return (
    <>
      <PageTitle title={t("navigation:page.settingsTitle")} />
      <div className="flex h-full w-full flex-col gap-6 overflow-y-auto px-4 py-4 md:px-8 md:py-5">
        <div className="flex flex-col gap-2">
          <button
            type="button"
            className="-ml-2 inline-flex min-h-11 w-fit items-center gap-1.5 rounded-fc-md px-2 text-fc-sm font-medium text-fc-fg-muted transition-colors hover:text-fc-fg"
            onClick={() =>
              navigate({
                to: "/dashboard/workspace/$workspaceId",
                params: { workspaceId: workspace?.id ?? "" },
              })
            }
          >
            <ChevronLeft className="size-4" />
            {t("navigation:page.backToWorkspace")}
          </button>
          <h1 className="text-2xl font-semibold text-fc-fg">
            {t("navigation:page.settingsTitle")}
          </h1>
        </div>

        <div className="grid gap-8 lg:grid-cols-[14rem_minmax(0,1fr)]">
          <nav
            aria-label={t("navigation:settingsLayout.sections")}
            className={cn(
              "flex flex-col gap-1 lg:sticky lg:top-6 lg:self-start",
              drilled && "max-lg:hidden",
            )}
          >
            {sections
              .filter((s) => !s.hidden)
              .map((s) => {
                const active = s.id === current?.id;
                return (
                  <Link
                    key={s.id}
                    to={s.to}
                    aria-current={active && drilled ? "page" : undefined}
                    className={cn(
                      row,
                      active
                        ? "text-fc-fg-muted hover:bg-fc-surface hover:text-fc-fg lg:bg-fc-accent lg:text-fc-accent-fg"
                        : "text-fc-fg-muted hover:bg-fc-surface hover:text-fc-fg",
                    )}
                  >
                    <s.icon className="size-[18px] shrink-0" />
                    <span className="min-w-0 flex-1 truncate">{s.label}</span>
                    <ChevronRight className="size-4 shrink-0 opacity-60 lg:hidden" />
                  </Link>
                );
              })}
          </nav>

          <div
            className={cn(
              "flex min-w-0 flex-col gap-10",
              !drilled && "max-lg:hidden",
            )}
          >
            <Link
              to={BASE}
              className="-mb-4 inline-flex min-h-11 w-fit items-center gap-1.5 rounded-fc-md text-fc-sm font-medium text-fc-fg-muted transition-colors hover:text-fc-fg lg:hidden"
            >
              <ChevronLeft className="size-4" />
              {t("navigation:settingsLayout.allSettings")}
            </Link>
            <Outlet />
          </div>
        </div>
      </div>
    </>
  );
}
