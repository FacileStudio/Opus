import type * as React from "react";

import { NavIdeas } from "@/components/nav-ideas";
import { NavMain } from "@/components/nav-main";
import { NavProjects } from "@/components/nav-projects";
import { SpaceSwitcher } from "@/components/space-switcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  useSidebar,
} from "@/components/ui/sidebar";
import { UserMenu } from "@/components/user-menu";
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
        <NavIdeas />
        <NavMain />
        <NavProjects />
      </SidebarContent>
      <SidebarFooter className="px-3 pt-3 pb-4">
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
