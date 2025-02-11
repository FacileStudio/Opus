import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/dashboard/workspace/$workspaceId/team/roles",
)({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/dashboard/workspace/$workspaceId/team/roles"!</div>;
}
