import useGetProject from "@/hooks/queries/project/use-get-project";
import useProjectStore from "@/store/project";
import { Link, Outlet, createFileRoute } from "@tanstack/react-router";
import { AlertCircle, LayoutGrid } from "lucide-react";
import { useEffect } from "react";

export const Route = createFileRoute(
  "/dashboard/workspace/$workspaceId/project/$projectId",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { workspaceId, projectId } = Route.useParams();
  const { data, isFetching, isError } = useGetProject({
    id: projectId,
    workspaceId,
  });
  const { setProject } = useProjectStore();

  useEffect(() => {
    if (data) {
      setProject(data);
    }
  }, [data, setProject]);

  if (isFetching) {
    return (
      <div className="flex w-full items-center justify-center h-screen flex-col md:flex-row bg-zinc-50 dark:bg-zinc-950">
        <div className="p-1.5 bg-linear-to-br from-indigo-500 to-purple-500 rounded-lg shadow-xs animate-spin">
          <LayoutGrid className="w-5 h-5 text-white" />
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex w-full items-center justify-center h-screen flex-col bg-zinc-50 dark:bg-zinc-950">
        <div className="w-14 h-14 mb-4 rounded-full bg-red-50 dark:bg-red-500/10 flex items-center justify-center">
          <AlertCircle className="w-7 h-7 text-red-500 dark:text-red-400" />
        </div>
        <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100 mb-2">
          Project not found
        </h1>
        <p className="text-zinc-500 dark:text-zinc-400 text-center max-w-md mb-6">
          The project you're looking for doesn't exist or you don't have access
          to it.
        </p>
        <Link
          to="/dashboard/workspace/$workspaceId"
          params={{ workspaceId }}
          className="px-4 py-2 bg-indigo-50 hover:bg-indigo-100 dark:bg-indigo-500/10 dark:hover:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-lg transition-colors font-medium inline-flex items-center gap-2"
        >
          Return to workspace
        </Link>
      </div>
    );
  }

  return <Outlet />;
}
