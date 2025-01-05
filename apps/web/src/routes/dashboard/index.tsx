import { Sidebar } from "@/components/common/sidebar";
import type { User } from "@/types/user";
import { redirect } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/dashboard/")({
  component: DashboardIndexRouteComponent,
  async beforeLoad({ context: { user, queryClient } }) {
    const queryData = await queryClient.ensureQueryData<{
      data: User;
    }>({
      queryKey: ["me"],
    });

    if (!user && !queryData?.data) {
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
