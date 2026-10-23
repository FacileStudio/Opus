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
    <Sidebar collapsible="offcanvas" variant="floating" {...props}>
      <SidebarHeader className="px-3 pt-4 pb-3">
        <SpaceSwitcher />
      </SidebarHeader>
      <SidebarContent className="overflow-hidden gap-1 px-1 py-1">
        <Search />
        <NavMain />
        <NavProjects />
      </SidebarContent>
      <SidebarFooter className="px-3 pt-3 pb-4">
        <div className="flex items-center gap-1">
          <div className="min-w-0 flex-1">
            <UserAvatar />
          </div>
          <ThemeToggleDropdown />
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
