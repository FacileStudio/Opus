import type * as React from "react";

import { NavMain } from "@/components/nav-main";
import { NavProjects } from "@/components/nav-projects";
import { SpaceSwitcher } from "@/components/space-switcher";
import { ThemeToggleDropdown } from "@/components/theme-toggle-dropdown";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  useSidebar,
} from "@/components/ui/sidebar";
import { UserAvatar } from "@/components/user-avatar";
import { VersionDisplay } from "@/components/version-display";
import { shortcuts } from "@/constants/shortcuts";
import { useRegisterShortcuts } from "@/hooks/use-keyboard-shortcuts";
import Search from "./search";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { toggleSidebar } = useSidebar();

  useRegisterShortcuts({
    modifierShortcuts: {
      [shortcuts.sidebar.prefix]: {
        [shortcuts.sidebar.toggle]: toggleSidebar,
      },
    },
  });

  return (
    <Sidebar
      collapsible="offcanvas"
      variant="inset"
      className="border-none py-3"
      {...props}
    >
      <SidebarHeader className="pt-2 pb-4">
        <SpaceSwitcher />
      </SidebarHeader>
      <SidebarContent className="overflow-hidden gap-1 py-1">
        <Search />
        <NavMain />
        <NavProjects />
      </SidebarContent>
      <SidebarFooter className="pt-4">
        <div className="flex h-11 items-center gap-2 rounded-md px-2 transition-colors hover:bg-sidebar-accent">
          <UserAvatar />
          <VersionDisplay />
          <div className="ms-auto">
            <ThemeToggleDropdown />
          </div>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
