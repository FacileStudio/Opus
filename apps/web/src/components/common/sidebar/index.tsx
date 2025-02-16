import { cn } from "@/lib/cn";
import { useUserPreferencesStore } from "@/store/user-preferences";
import { SidebarContent } from "./sidebar-content";
import SidebarFooter from "./sidebar-footer";
import { SidebarHeader } from "./sidebar-header";

export function Sidebar() {
  const { isSidebarOpened, setIsSidebarOpened } = useUserPreferencesStore();

  return (
    <>
      <div
        onClick={() => setIsSidebarOpened()}
        onKeyDown={(e) => {
          if (e.key === "Escape") {
            setIsSidebarOpened();
          }
        }}
        className={cn(
          "fixed inset-0 bg-black/50 z-30 lg:hidden",
          "h-[100dvh]",
          isSidebarOpened ? "opacity-100" : "opacity-0 pointer-events-none",
          "transition-opacity duration-300 ease-in-out",
        )}
      />

      <div className="w-16 shrink-0 lg:hidden" />

      <div
        className={cn(
          "flex bg-white dark:bg-zinc-900 border-r border-zinc-200 dark:border-zinc-800 h-screen flex-col",
          "transition-all duration-300 ease-in-out",
          "min-h-screen",
          "fixed lg:relative top-0 left-0",
          "z-30",
          !isSidebarOpened ? "w-16" : "w-64",
        )}
      >
        <SidebarHeader />
        <SidebarContent />
        <SidebarFooter />
      </div>
    </>
  );
}
