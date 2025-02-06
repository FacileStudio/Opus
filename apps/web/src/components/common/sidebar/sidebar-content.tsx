import { cn } from "@/lib/cn";
import Workspaces from "./sections/workspaces";

export function SidebarContent() {
  return (
    <nav className="flex-1 overflow-y-auto p-3">
      <div className={cn("space-y-4")}>
        <Workspaces />
      </div>
    </nav>
  );
}
