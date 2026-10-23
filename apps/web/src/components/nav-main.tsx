import {
  AltArrowRight,
  ChecklistMinimalistic,
  Folder,
  Letter,
  UsersGroupTwoRounded,
} from "@solar-icons/react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import {
  Collapsible,
  CollapsiblePanel,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { usePendingInvitations } from "@/hooks/queries/invitation/use-pending-invitations";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";

export const navGroupLabelClassName =
  "cursor-pointer justify-between px-3 text-xs font-semibold uppercase tracking-wide text-sidebar-foreground/70 outline-hidden transition-colors hover:text-foreground";

export const navRowClassName =
  "h-11 rounded-md px-3 text-sm text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground active:bg-sidebar-accent data-[active=true]:bg-sidebar-primary data-[active=true]:text-sidebar-primary-foreground";

export function NavMain() {
  const { t } = useTranslation();
  const { data: workspace } = useActiveWorkspace();
  const navigate = useNavigate();
  const { data: invitations = [] } = usePendingInvitations();

  if (!workspace) return null;

  const pendingCount = invitations.length;

  const navItems = [
    {
      title: t("navigation:sidebar.projects"),
      url: `/dashboard/workspace/${workspace.id}`,
      icon: Folder,
      isActive:
        window.location.pathname === `/dashboard/workspace/${workspace.id}`,
      badge: null,
    },
    {
      title: t("navigation:sidebar.myTasks"),
      url: `/dashboard/workspace/${workspace.id}/my-tasks`,
      icon: ChecklistMinimalistic,
      isActive:
        window.location.pathname ===
        `/dashboard/workspace/${workspace.id}/my-tasks`,
      badge: null,
    },
    {
      title: t("navigation:sidebar.members"),
      url: `/dashboard/workspace/${workspace.id}/members`,
      icon: UsersGroupTwoRounded,
      isActive:
        window.location.pathname ===
        `/dashboard/workspace/${workspace.id}/members`,
      badge: null,
    },
    {
      title: t("navigation:sidebar.invitations"),
      url: "/dashboard/invitations",
      icon: Letter,
      isActive: window.location.pathname === "/dashboard/invitations",
      badge: pendingCount > 0 ? pendingCount : null,
    },
  ];

  return (
    <Collapsible defaultOpen className="group/collapsible">
      <SidebarGroup className="gap-1 p-2">
        <CollapsibleTrigger
          className="data-panel-open:[&_svg]:rotate-90"
          render={
            <SidebarGroupLabel
              className={navGroupLabelClassName}
              render={<button type="button" />}
            />
          }
        >
          <span>{t("navigation:sidebar.overview")}</span>
          <AltArrowRight
            size={14}
            className="transition-transform duration-200"
          />
        </CollapsibleTrigger>
        <CollapsiblePanel>
          <SidebarGroupContent>
            <SidebarMenu className="gap-1">
              {navItems.map((item) => (
                <SidebarMenuItem key={item.url}>
                  <SidebarMenuButton
                    tooltip={item.title}
                    isActive={item.isActive}
                    size="default"
                    className={navRowClassName}
                    onClick={() => navigate({ to: item.url })}
                  >
                    <item.icon className="size-4.5 shrink-0" />
                    <span className="min-w-0 flex-1 truncate">
                      {item.title}
                    </span>
                    {item.badge !== null && (
                      <span
                        className={
                          item.isActive
                            ? "flex h-5 min-w-5 items-center justify-center rounded-full bg-sidebar-primary-foreground/20 px-1 text-[11px] font-medium text-sidebar-primary-foreground"
                            : "flex h-5 min-w-5 items-center justify-center rounded-full bg-sidebar-accent px-1 text-[11px] font-medium text-sidebar-foreground"
                        }
                      >
                        {item.badge}
                      </span>
                    )}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </CollapsiblePanel>
      </SidebarGroup>
    </Collapsible>
  );
}
