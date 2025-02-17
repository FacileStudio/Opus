import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/dashboard/teams/$workspaceId/_layout/roles",
)({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/dashboard/workspace/$workspaceId/team/roles"!</div>;
}
