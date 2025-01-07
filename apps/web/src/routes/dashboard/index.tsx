import { Sidebar } from "@/components/common/sidebar";
import { redirect } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/dashboard/")({
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
  return (
    <>
      <Sidebar />
      <main className="flex-1 overflow-hidden p-6" />
    </>
  );
}
