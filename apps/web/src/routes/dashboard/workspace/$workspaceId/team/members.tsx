import TeamTable from "@/components/team/team-table";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/dashboard/workspace/$workspaceId/team/members",
)({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <>
      <div className="flex-1 p-6">
        <div className="w-full">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
              Team Members
            </h1>
          </div>
        </div>
        <div className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
          <TeamTable />
        </div>
      </div>
    </>
  );
}
