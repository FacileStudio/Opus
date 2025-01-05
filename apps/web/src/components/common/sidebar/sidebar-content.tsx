import Team from "./sections/team";
import Workspaces from "./sections/workspaces";

export function SidebarContent() {
  return (
    <nav className="flex-1 overflow-y-auto p-3">
      <div className="space-y-4">
        <Workspaces />
        <Team />
      </div>
    </nav>
  );
}
