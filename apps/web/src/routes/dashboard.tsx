import { Sidebar } from "@/components/common/sidebar";
import EmptyWorkspaceState from "@/components/workspace/empty-state";
import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspaces";
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
  const { data } = useGetWorkspaces();

  const isWorkspacesEmpty = data && data.length === 0;

  return (
    <>
      <Sidebar />
      <main className="flex-1 overflow-hidden scroll-smooth">
        {isWorkspacesEmpty ? <EmptyWorkspaceState /> : <Outlet />}
      </main>
    </>
  );
}
