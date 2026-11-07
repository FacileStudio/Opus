import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace",
)({
  component: RouteComponent,
});

function RouteComponent() {
  return <Outlet />;
}
