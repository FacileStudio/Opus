import { AltArrowDown as ChevronDown } from "@solar-icons/react";
import { useNavigate } from "@tanstack/react-router";
import * as React from "react";
import { useTranslation } from "react-i18next";

import { SpaceIcon } from "@/components/icons/space-icon";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { shortcuts } from "@/constants/shortcuts";
import { getApiUrl } from "@/fetchers/get-api-url";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspaces";
import {
  getModifierKeyText,
  useRegisterShortcuts,
} from "@/hooks/use-keyboard-shortcuts";
import { authClient } from "@/lib/auth-client";
import type { Workspace } from "@/types/workspace";
import CreateSpaceModal from "./shared/modals/create-space-modal";

export function SpaceSwitcher() {
  const { t } = useTranslation();
  const { data: workspace } = useActiveWorkspace();
  const { data: workspaces } = useGetWorkspaces();
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = React.useState(false);
  const [isCreateSpaceModalOpen, setIsCreateSpaceModalOpen] =
    React.useState(false);
  const [isSwitching, setIsSwitching] = React.useState(false);

  const handleWorkspaceChange = React.useCallback(
    async (selectedWorkspace: Workspace) => {
      if (isSwitching) return;

      setIsSwitching(true);
      try {
        await authClient.organization.setActive({
          organizationId: selectedWorkspace.id,
        });

        setTimeout(() => {
          navigate({
            to: "/dashboard/workspace/$workspaceId",
            params: { workspaceId: selectedWorkspace.id },
          });
        }, 50);
      } catch (error) {
        console.error("Failed to switch workspace:", error);
      } finally {
        setTimeout(() => setIsSwitching(false), 100);
      }
    },
    [navigate, isSwitching],
  );

  React.useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (!workspaces || workspaces.length === 0) return;

      if (
        (event.metaKey || event.ctrlKey) &&
        event.key >= "1" &&
        event.key <= "9"
      ) {
        event.preventDefault();
        const index = Number.parseInt(event.key, 10) - 1;
        if (index < workspaces.length) {
          handleWorkspaceChange(workspaces[index]);
          setIsOpen(false);
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, workspaces, handleWorkspaceChange]);

  useRegisterShortcuts({
    sequentialShortcuts: {
      [shortcuts.workspace.prefix]: {
        [shortcuts.workspace.switch]: () => {
          setIsOpen(true);
        },
        [shortcuts.workspace.create]: () => {
          setIsCreateSpaceModalOpen(true);
        },
      },
    },
  });

  if (!workspace) {
    return null;
  }

  return (
    <>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
            <DropdownMenuTrigger
              render={
                <SidebarMenuButton
                  className="group h-11 w-full rounded-md bg-sidebar-accent/50 px-3 transition-colors hover:bg-sidebar-accent"
                  size="default"
                />
              }
            >
              <div className="flex items-center gap-2 min-w-0 w-full">
                <Avatar className="size-5 shrink-0 rounded-md">
                  {workspace.logo ? (
                    <AvatarImage
                      src={getApiUrl(
                        `/workspace/${workspace.id}/logo?v=${workspace.logo}`,
                      )}
                      alt={workspace.name}
                    />
                  ) : null}
                  <AvatarFallback className="rounded-md">
                    <SpaceIcon className="size-3.5" />
                  </AvatarFallback>
                </Avatar>
                <span
                  className={`truncate text-sm font-medium text-foreground ${isSwitching ? "opacity-50" : ""}`}
                >
                  {workspace.name}
                </span>
              </div>
              <ChevronDown
                size={14}
                className={`ml-1 text-foreground/70 opacity-90 group-hover:opacity-100 data-[state=open]:opacity-100 data-[state=open]:rotate-180 transition-all duration-200 ease-out ${isSwitching ? "animate-spin" : ""}`}
                data-state={isOpen ? "open" : "closed"}
              />
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="min-w-(--anchor-width) text-sidebar-foreground"
              align="start"
              side="bottom"
              sideOffset={4}
            >
              <DropdownMenuGroup>
                <DropdownMenuLabel>
                  {t("navigation:workspaceSwitcher.workspaces")}
                </DropdownMenuLabel>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />

              {workspaces?.map((ws: Workspace, index: number) => (
                <DropdownMenuItem
                  key={ws.id}
                  onClick={() => {
                    if (!isSwitching && ws.id !== workspace.id) {
                      handleWorkspaceChange(ws);
                      setIsOpen(false);
                    }
                  }}
                  disabled={isSwitching || ws.id === workspace.id}
                  className="h-7 text-sm data-highlighted:bg-sidebar-accent data-highlighted:text-sidebar-accent-foreground"
                >
                  <span className="flex-1 text-left">
                    {isSwitching && ws.id === workspace?.id
                      ? t("navigation:workspaceSwitcher.switching")
                      : ws.name}
                  </span>
                  <DropdownMenuShortcut>
                    {getModifierKeyText()} {index > 8 ? "0" : index + 1}
                  </DropdownMenuShortcut>
                </DropdownMenuItem>
              ))}

              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => {
                  setIsCreateSpaceModalOpen(true);
                  setIsOpen(false);
                }}
                className="h-7 text-sm data-highlighted:bg-sidebar-accent data-highlighted:text-sidebar-accent-foreground"
              >
                <span>{t("navigation:workspaceSwitcher.addWorkspace")}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>

      <CreateSpaceModal
        open={isCreateSpaceModalOpen}
        onClose={() => setIsCreateSpaceModalOpen(false)}
      />
    </>
  );
}
