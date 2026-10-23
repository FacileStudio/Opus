import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import getTasks from "@/fetchers/task/get-tasks";
import useGetProjects from "@/hooks/queries/project/use-get-projects";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/task/$key",
)({
  component: RouteComponent,
});

function parseTaskKey(key: string) {
  const separator = key.lastIndexOf("-");
  const number = Number(key.slice(separator + 1));
  if (separator < 1 || !Number.isInteger(number)) return null;
  return { slug: key.slice(0, separator).toLowerCase(), number };
}

function RouteComponent() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { workspaceId, key } = Route.useParams();
  const { data: projects, isLoading } = useGetProjects({ workspaceId });
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (isLoading) return;
    const parsed = parseTaskKey(key);
    const project = projects?.find(
      (candidate) => candidate.slug?.toLowerCase() === parsed?.slug,
    );
    if (!parsed || !project) {
      setNotFound(true);
      return;
    }

    let cancelled = false;
    getTasks(String(project.id))
      .then((data) => {
        const task = [
          ...data.columns.flatMap((column) => column.tasks),
          ...data.plannedTasks,
          ...data.archivedTasks,
        ].find((candidate) => candidate.number === parsed.number);
        if (cancelled) return;
        if (!task) {
          setNotFound(true);
          return;
        }
        navigate({
          to: "/dashboard/workspace/$workspaceId/project/$projectId/task/$taskId",
          params: {
            workspaceId,
            projectId: String(project.id),
            taskId: String(task.id),
          },
          replace: true,
        });
      })
      .catch(() => {
        if (!cancelled) setNotFound(true);
      });

    return () => {
      cancelled = true;
    };
  }, [isLoading, projects, key, workspaceId, navigate]);

  return (
    <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">
      {notFound
        ? t("tasks:shortLink.notFound", { key })
        : t("tasks:shortLink.loading")}
    </div>
  );
}
