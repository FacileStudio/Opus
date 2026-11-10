import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { IdeasIcon } from "@/components/ideas/ideas-icon";
import { navRowClassName } from "@/components/nav-main";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useIdeas } from "@/hooks/queries/ideas/use-ideas";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { countNewIdeas, ideasScope } from "@/lib/ideas";

export function NavIdeas() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: workspace, isSettled } = useActiveWorkspace();
  const scope = ideasScope(workspace?.id, isSettled);
  const { data } = useIdeas(scope.workspaceId, scope.ready);

  if (!scope.ready) return null;

  const isActive = window.location.pathname === "/ideas";
  const count = data?.ai_access ? countNewIdeas(data.ideas) : 0;
  const title = t("navigation:sidebar.ideas");

  return (
    <SidebarGroup className="gap-1 p-2 pb-0">
      <SidebarGroupContent>
        <SidebarMenu className="gap-1">
          <SidebarMenuItem>
            <SidebarMenuButton
              tooltip={title}
              isActive={isActive}
              size="default"
              className={navRowClassName}
              onClick={() => navigate({ to: "/ideas" })}
              data-testid="nav-ideas"
            >
              <IdeasIcon className="size-4.5 shrink-0" />
              <span className="min-w-0 flex-1 truncate">{title}</span>
              {count > 0 && (
                <span
                  role="status"
                  aria-label={t("ideas:newCount", { count })}
                  className={
                    isActive
                      ? "flex h-5 min-w-5 shrink-0 items-center justify-center whitespace-nowrap rounded-full bg-sidebar-primary-foreground/20 px-1 text-[11px] font-medium text-sidebar-primary-foreground"
                      : "flex h-5 min-w-5 shrink-0 items-center justify-center whitespace-nowrap rounded-full bg-sidebar-accent px-1 text-[11px] font-medium text-sidebar-foreground"
                  }
                >
                  {count}
                </span>
              )}
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
