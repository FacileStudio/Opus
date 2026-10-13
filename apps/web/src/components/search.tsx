"use client";

import { Magnifier as SearchIcon } from "@solar-icons/react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import SearchCommandMenu from "@/components/search-command-menu";
import { SidebarGroup } from "@/components/ui/sidebar";
import { shortcuts } from "@/constants/shortcuts";
import { useRegisterShortcuts } from "@/hooks/use-keyboard-shortcuts";

export default function Search() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  useRegisterShortcuts({
    shortcuts: {
      [shortcuts.search.prefix]: () => {
        setOpen(true);
      },
    },
  });

  return (
    <SidebarGroup>
      <button
        className="flex h-11 w-full cursor-pointer items-center gap-2.5 rounded-md px-3 text-sm text-sidebar-foreground outline-hidden transition-colors hover:bg-sidebar-accent hover:text-foreground focus-visible:ring-2 focus-visible:ring-sidebar-ring active:bg-sidebar-accent"
        onClick={() => setOpen(true)}
        type="button"
      >
        <SearchIcon aria-hidden="true" className="shrink-0" size={18} />
        <span className="min-w-0 flex-1 truncate text-left">
          {t("navigation:commandPalette.search")}
        </span>
        <kbd className="shrink-0 text-xs opacity-50">
          {shortcuts.search.prefix}
        </kbd>
      </button>

      <SearchCommandMenu open={open} setOpen={setOpen} />
    </SidebarGroup>
  );
}
