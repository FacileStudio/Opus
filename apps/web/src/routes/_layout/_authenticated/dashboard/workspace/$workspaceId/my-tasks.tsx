import { ChecklistMinimalistic } from "@solar-icons/react";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { format } from "date-fns";
import { useTranslation } from "react-i18next";
import WorkspaceLayout from "@/components/common/workspace-layout";
import PageTitle from "@/components/page-title";
import SubtaskProgress from "@/components/task/subtask-progress";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import useGetMyTasks from "@/hooks/queries/task/use-get-my-tasks";
import { useWorkspaceWebSocket } from "@/hooks/use-project-websocket";
import { getColumnIcon } from "@/lib/column";
import { getPriorityIcon } from "@/lib/priority";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/my-tasks",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { workspaceId } = Route.useParams();
  const { data: tasks = [], isLoading } = useGetMyTasks(workspaceId);
  useWorkspaceWebSocket(workspaceId);

  const byProject = new Map<string, typeof tasks>();
  for (const task of tasks) {
    byProject.set(task.projectId, [
      ...(byProject.get(task.projectId) ?? []),
      task,
    ]);
  }
  const groups = [...byProject.values()];

  return (
    <>
      <PageTitle title={t("tasks:myTasks.pageTitle")} />
      <WorkspaceLayout title={t("tasks:myTasks.pageTitle")}>
        {!isLoading && groups.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ChecklistMinimalistic />
              </EmptyMedia>
              <EmptyTitle>{t("tasks:myTasks.emptyTitle")}</EmptyTitle>
              <EmptyDescription>
                {t("tasks:myTasks.emptyDescription")}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="flex flex-col gap-4 p-4">
            {groups.map((group) => (
              <section
                key={group[0].projectId}
                className="rounded-lg bg-card p-2"
              >
                <h2 className="px-2 py-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {group[0].projectName}
                </h2>
                {group.map((task) => (
                  <button
                    key={task.id}
                    type="button"
                    className="flex w-full items-center gap-3 rounded-md px-2 py-2 text-left text-sm outline-none transition-colors hover:bg-muted"
                    onClick={() =>
                      navigate({
                        to: `/dashboard/workspace/${workspaceId}/project/${task.projectId}/task/${task.id}`,
                      })
                    }
                  >
                    <span className="w-20 shrink-0 text-xs text-muted-foreground">
                      {task.projectSlug.toUpperCase()}-{task.number}
                    </span>
                    <span className="min-w-0 flex-1 truncate">
                      {task.title}
                    </span>
                    <SubtaskProgress
                      done={task.subtaskDone}
                      total={task.subtaskTotal}
                    />
                    <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                      {getColumnIcon(task.status, false)}
                      {t(`tasks:status.${task.status}`, {
                        defaultValue: task.status,
                      })}
                    </span>
                    <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                      {getPriorityIcon(task.priority)}
                      {t(`tasks:priority.${task.priority}`, {
                        defaultValue: task.priority,
                      })}
                    </span>
                    <span className="w-14 shrink-0 text-right text-xs text-muted-foreground">
                      {task.dueDate
                        ? format(new Date(task.dueDate), "MMM d")
                        : ""}
                    </span>
                  </button>
                ))}
              </section>
            ))}
          </div>
        )}
      </WorkspaceLayout>
    </>
  );
}
