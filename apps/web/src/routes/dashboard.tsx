import { Sidebar } from "@/components/common/sidebar";
import SelectWorkspaceState from "@/components/workspace/select-workspace-state";
import useWorkspaceStore from "@/store/workspace";
import { Outlet, createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/dashboard")({
  component: DashboardIndexRouteComponent,
  async beforeLoad({ context: { user } }) {
    if (user === null) {
      throw redirect({
        to: "/auth/sign-in",
      });
    }
  },
});

function DashboardIndexRouteComponent() {
  const { workspace } = useWorkspaceStore();

  return (
    <>
      <Sidebar />
      <main className="flex-1 overflow-hidden scroll-smooth">
        {workspace ? <Outlet /> : <SelectWorkspaceState />}
      </main>
    </>
  );
}
