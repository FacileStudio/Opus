import { ChecklistMinimalistic, List, Widget2 } from "@solar-icons/react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { differenceInCalendarDays, format } from "date-fns";
import type { ReactNode } from "react";
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
import { Skeleton } from "@/components/ui/skeleton";
import icons from "@/constants/project-icons";
import useGetMyTasks from "@/hooks/queries/task/use-get-my-tasks";
import { useWorkspaceWebSocket } from "@/hooks/use-project-websocket";
import { cn } from "@/lib/cn";
import { getColumnIcon } from "@/lib/column";
import { getPriorityIcon } from "@/lib/priority";
import { useUserPreferencesStore } from "@/store/user-preferences";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/my-tasks",
)({
  component: RouteComponent,
});

type MyTask = NonNullable<ReturnType<typeof useGetMyTasks>["data"]>[number];

const statusOrder = ["planned", "to-do", "in-progress", "in-review"];
const priorityOrder = ["urgent", "high", "medium", "low", "no-priority"];

function daysUntilDue(task: MyTask) {
  return task.dueDate
    ? differenceInCalendarDays(new Date(task.dueDate), new Date())
    : null;
}

function compareTasks(a: MyTask, b: MyTask) {
  const dueA = daysUntilDue(a) ?? Number.POSITIVE_INFINITY;
  const dueB = daysUntilDue(b) ?? Number.POSITIVE_INFINITY;
  if (dueA !== dueB) return dueA - dueB;
  const rank = (task: MyTask) => {
    const index = priorityOrder.indexOf(task.priority);
    return index === -1 ? priorityOrder.length : index;
  };
  return rank(a) - rank(b) || a.number - b.number;
}

function groupBy(tasks: MyTask[], key: (task: MyTask) => string) {
  const groups = new Map<string, MyTask[]>();
  for (const task of tasks) {
    groups.set(key(task), [...(groups.get(key(task)) ?? []), task]);
  }
  return [...groups.entries()];
}

function compareStatuses(a: string, b: string) {
  const rank = (status: string) => {
    const index = statusOrder.indexOf(status);
    return index === -1 ? statusOrder.length : index;
  };
  return rank(a) - rank(b) || a.localeCompare(b);
}

function RouteComponent() {
  const { t } = useTranslation();
  const { workspaceId } = Route.useParams();
  const { data = [], isLoading } = useGetMyTasks(workspaceId);
  const view = useUserPreferencesStore((s) => s.myTasksView);
  const setView = useUserPreferencesStore((s) => s.setMyTasksView);
  useWorkspaceWebSocket(workspaceId);

  const tasks = [...data].sort(compareTasks);
  const overdue = tasks.filter((task) => (daysUntilDue(task) ?? 0) < 0).length;
  const dueSoon = tasks.filter((task) => {
    const days = daysUntilDue(task);
    return days !== null && days >= 0 && days <= 7;
  }).length;

  return (
    <>
      <PageTitle title={t("tasks:myTasks.pageTitle")} />
      <WorkspaceLayout
        title={t("tasks:myTasks.pageTitle")}
        headerActions={<ViewToggle view={view} onChange={setView} />}
      >
        <div
          className={cn(
            "flex w-full flex-col gap-10 px-4 py-6 md:px-6",
            view === "list" && "mx-auto max-w-5xl",
          )}
        >
          <header className="flex flex-col gap-1">
            <h1 className="text-2xl font-semibold tracking-tight">
              {t("tasks:myTasks.pageTitle")}
            </h1>
            {isLoading ? (
              <Skeleton className="h-5 w-64" />
            ) : (
              <p className="text-sm text-muted-foreground">
                {t("tasks:myTasks.summary", { count: tasks.length })}
                {overdue > 0 && (
                  <span className="text-destructive-foreground">
                    {" · "}
                    {t("tasks:myTasks.overdueCount", { count: overdue })}
                  </span>
                )}
                {dueSoon > 0 &&
                  ` · ${t("tasks:myTasks.dueSoonCount", { count: dueSoon })}`}
              </p>
            )}
          </header>

          {isLoading ? (
            <LoadingState />
          ) : tasks.length === 0 ? (
            <Empty className="rounded-xl bg-card py-12">
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
          ) : view === "columns" ? (
            <ColumnsView tasks={tasks} workspaceId={workspaceId} />
          ) : (
            <ListView tasks={tasks} workspaceId={workspaceId} />
          )}
        </div>
      </WorkspaceLayout>
    </>
  );
}

type ViewToggleProps = {
  view: "list" | "columns";
  onChange: (view: "list" | "columns") => void;
};

function ViewToggle({ view, onChange }: ViewToggleProps) {
  const { t } = useTranslation();
  const options = [
    { value: "list", label: t("tasks:myTasks.viewList"), icon: <List /> },
    {
      value: "columns",
      label: t("tasks:myTasks.viewColumns"),
      icon: <Widget2 />,
    },
  ] as const;

  return (
    <fieldset
      aria-label={t("tasks:myTasks.viewLabel")}
      className="inline-flex items-center gap-0.5 rounded-full border-0 bg-muted p-0.5"
    >
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          aria-pressed={view === option.value}
          onClick={() => onChange(option.value)}
          className={cn(
            "inline-flex h-7 items-center gap-1.5 rounded-full px-3 text-xs font-medium transition-colors [&_svg]:size-3.5",
            view === option.value
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          {option.icon}
          <span className="hidden sm:inline">{option.label}</span>
        </button>
      ))}
    </fieldset>
  );
}

function LoadingState() {
  return (
    <div className="flex flex-col gap-4">
      {[0, 1].map((group) => (
        <div key={group} className="flex flex-col gap-2 rounded-xl bg-card p-5">
          <Skeleton className="h-5 w-40" />
          {[0, 1, 2].map((row) => (
            <Skeleton key={row} className="h-9 w-full" />
          ))}
        </div>
      ))}
    </div>
  );
}

type ViewProps = {
  tasks: MyTask[];
  workspaceId: string;
};

function ProjectIcon({
  icon,
  className,
}: {
  icon: string;
  className?: string;
}) {
  const Icon = icons[icon as keyof typeof icons] || icons.Layout;
  return <Icon className={cn("shrink-0 text-muted-foreground", className)} />;
}

function ListView({ tasks, workspaceId }: ViewProps) {
  return (
    <div className="flex flex-col gap-10">
      {groupBy(tasks, (task) => String(task.projectId)).map(
        ([projectId, group]) => (
          <section key={projectId} className="flex flex-col gap-4">
            <div className="flex items-center gap-2">
              <ProjectIcon icon={group[0].projectIcon} className="size-4.5" />
              <h2 className="text-lg font-semibold tracking-tight">
                {group[0].projectName}
              </h2>
              <span className="rounded-full bg-muted px-2 text-xs leading-5 text-muted-foreground tabular-nums">
                {group.length}
              </span>
            </div>
            <ul className="overflow-hidden rounded-xl bg-card">
              {group.map((task) => (
                <li
                  key={task.id}
                  className="border-border/60 border-b last:border-b-0"
                >
                  <TaskLink task={task} workspaceId={workspaceId}>
                    <TaskRow task={task} />
                  </TaskLink>
                </li>
              ))}
            </ul>
          </section>
        ),
      )}
    </div>
  );
}

function ColumnsView({ tasks, workspaceId }: ViewProps) {
  const { t } = useTranslation();
  const columns = groupBy(tasks, (task) => task.status).sort(([a], [b]) =>
    compareStatuses(a, b),
  );

  return (
    <div className="-mx-4 flex snap-x gap-4 overflow-x-auto px-4 pb-2 md:-mx-6 md:px-6">
      {columns.map(([status, group]) => (
        <section
          key={status}
          className="flex w-[85vw] max-w-80 shrink-0 snap-start flex-col gap-3 rounded-xl bg-card p-3 sm:w-80"
        >
          <div className="flex items-center gap-2 px-1">
            {getColumnIcon(status, false)}
            <h2 className="text-sm font-semibold">
              {t(`tasks:status.${status}`, { defaultValue: status })}
            </h2>
            <span className="rounded-full bg-muted px-2 text-xs leading-5 text-muted-foreground tabular-nums">
              {group.length}
            </span>
          </div>
          <ul className="flex flex-col gap-2">
            {group.map((task) => (
              <li key={task.id}>
                <TaskLink
                  task={task}
                  workspaceId={workspaceId}
                  className="rounded-lg bg-background p-3 shadow-xs/5 hover:bg-muted/60"
                >
                  <TaskCard task={task} />
                </TaskLink>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}

type TaskLinkProps = {
  task: MyTask;
  workspaceId: string;
  className?: string;
  children: ReactNode;
};

function TaskLink({ task, workspaceId, className, children }: TaskLinkProps) {
  return (
    <Link
      to="/dashboard/workspace/$workspaceId/project/$projectId/task/$taskId"
      params={{
        workspaceId,
        projectId: String(task.projectId),
        taskId: String(task.id),
      }}
      className={cn(
        "block outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring",
        className ?? "hover:bg-muted/60",
      )}
    >
      {children}
    </Link>
  );
}

function TaskKey({ task }: { task: MyTask }) {
  return (
    <span className="shrink-0 font-mono text-xs text-muted-foreground tabular-nums">
      {task.projectSlug.toUpperCase()}-{task.number}
    </span>
  );
}

function TaskRow({ task }: { task: MyTask }) {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-11 items-center gap-3 px-4 py-2 text-sm">
      <span
        className="shrink-0"
        title={t(`tasks:status.${task.status}`, { defaultValue: task.status })}
      >
        {getColumnIcon(task.status, false)}
      </span>
      <TaskKey task={task} />
      <span className="min-w-0 flex-1 truncate">{task.title}</span>
      <SubtaskProgress done={task.subtaskDone} total={task.subtaskTotal} />
      <span
        className="flex shrink-0 items-center"
        title={t(`tasks:priority.${task.priority}`, {
          defaultValue: task.priority,
        })}
      >
        {getPriorityIcon(task.priority)}
      </span>
      <DueDate task={task} className="w-16 justify-end" />
    </div>
  );
}

function TaskCard({ task }: { task: MyTask }) {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col gap-2 text-sm">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <ProjectIcon icon={task.projectIcon} className="size-3.5" />
        <span className="truncate">{task.projectName}</span>
        <span className="ml-auto">
          <TaskKey task={task} />
        </span>
      </div>
      <p className="line-clamp-2 font-medium">{task.title}</p>
      <div className="flex items-center gap-3">
        <span
          className="flex items-center gap-1 text-xs text-muted-foreground"
          title={t("tasks:priority.label")}
        >
          {getPriorityIcon(task.priority)}
          {t(`tasks:priority.${task.priority}`, {
            defaultValue: task.priority,
          })}
        </span>
        <SubtaskProgress done={task.subtaskDone} total={task.subtaskTotal} />
        <DueDate task={task} className="ml-auto" />
      </div>
    </div>
  );
}

function DueDate({ task, className }: { task: MyTask; className?: string }) {
  const { t } = useTranslation();
  const days = daysUntilDue(task);
  if (days === null || !task.dueDate) {
    return <span className={className} />;
  }

  const label =
    days === 0
      ? t("tasks:myTasks.dueToday")
      : days === 1
        ? t("tasks:myTasks.dueTomorrow")
        : format(new Date(task.dueDate), "MMM d");

  return (
    <span
      className={cn(
        "flex shrink-0 items-center text-xs tabular-nums",
        days < 0 && "font-medium text-destructive-foreground",
        days >= 0 && days <= 1 && "font-medium text-warning-foreground",
        days > 1 && "text-muted-foreground",
        className,
      )}
    >
      {label}
    </span>
  );
}
