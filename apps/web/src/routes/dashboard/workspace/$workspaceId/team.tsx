import { cn } from "@/lib/cn";
import {
  Link,
  Outlet,
  createFileRoute,
  useRouterState,
} from "@tanstack/react-router";
import { Shield, UserPlus, Users } from "lucide-react";

export const Route = createFileRoute("/dashboard/workspace/$workspaceId/team")({
  component: RouteComponent,
});

function RouteComponent() {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { workspaceId } = Route.useParams();

  return (
    <div className="flex-1 flex h-screen">
      <aside className="w-64 border-r border-zinc-200 dark:border-zinc-800 p-6">
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
          Team
        </h2>
        <nav className="space-y-1">
          <Link
            className={`flex items-center px-3 py-2 rounded-lg  text-sm ${
              pathname.includes("/members")
                ? "bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400"
                : "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800"
            }`}
            to="/dashboard/workspace/$workspaceId/team/members"
            params={{
              workspaceId,
            }}
          >
            <Users className="w-4 h-4 mr-3" />
            Members
          </Link>

          <Link
            className={`flex items-center px-3 py-2 rounded-lg  text-sm ${
              pathname.includes("/invitations")
                ? "bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400"
                : "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800"
            }`}
            to="/dashboard/workspace/$workspaceId/team/invitations"
            params={{
              workspaceId,
            }}
          >
            <UserPlus className="w-4 h-4 mr-3" />
            Invitations
          </Link>

          <div
            className={cn(
              "flex items-center px-3 py-2 rounded-lg text-sm opacity-50 cursor-not-allowed",
              "text-zinc-400 dark:text-zinc-600",
            )}
          >
            <Shield className="w-4 h-4 mr-3" />
            <span className="flex items-center gap-2">
              Roles
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
                Soon
              </span>
            </span>
          </div>
        </nav>
      </aside>
      <Outlet />
    </div>
  );
}
