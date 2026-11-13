import {
  CalendarDate,
  CheckRead,
  HamburgerMenu,
  Widget5,
} from "@solar-icons/react";
import { useTranslation } from "react-i18next";
import type { ProjectView } from "@/components/common/project-layout";
import { MdiPlus as AddCircle } from "@/components/icons/mdi-plus";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import icons from "@/constants/project-icons";
import useGetProjects from "@/hooks/queries/project/use-get-projects";
import { cn } from "@/lib/cn";

type MobileProjectNavProps = {
  workspaceId: string;
  projectId: string;
  activeView: ProjectView;
  onSelectBoard: () => void;
  onSelectBacklog: () => void;
  onSelectGantt: () => void;
  onSelectIssues?: () => void;
  onSelectProject: (projectId: string) => void;
  onAddProject: () => void;
};

export default function MobileProjectNav({
  workspaceId,
  projectId,
  activeView,
  onSelectBoard,
  onSelectBacklog,
  onSelectGantt,
  onSelectIssues,
  onSelectProject,
  onAddProject,
}: MobileProjectNavProps) {
  const { t } = useTranslation();
  const { data: projects = [] } = useGetProjects({ workspaceId });

  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button
            variant="ghost"
            size="icon-xs"
            className="size-7 border border-transparent"
          />
        }
      >
        <HamburgerMenu size={16} />
      </PopoverTrigger>
      <PopoverContent align="start" className="w-72 p-2">
        <div className="space-y-3">
          <div className="space-y-1">
            <p className="px-1 text-[11px] font-medium tracking-wide text-muted-foreground uppercase">
              View
            </p>
            <div
              className={cn(
                "grid gap-1",
                onSelectIssues ? "grid-cols-4" : "grid-cols-3",
              )}
            >
              <button
                type="button"
                onClick={onSelectBacklog}
                className={cn(
                  "flex w-full items-center justify-center gap-1 whitespace-nowrap rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                  activeView === "backlog"
                    ? "border-border bg-secondary text-foreground"
                    : "border-transparent text-muted-foreground hover:bg-accent",
                )}
              >
                Backlog
              </button>
              <button
                type="button"
                onClick={onSelectBoard}
                className={cn(
                  "flex w-full items-center justify-center gap-1 whitespace-nowrap rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                  activeView === "board"
                    ? "border-border bg-secondary text-foreground"
                    : "border-transparent text-muted-foreground hover:bg-accent",
                )}
              >
                <Widget5 size={14} />
                Board
              </button>
              <button
                type="button"
                onClick={onSelectGantt}
                className={cn(
                  "flex w-full items-center justify-center gap-1 whitespace-nowrap rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                  activeView === "gantt"
                    ? "border-border bg-secondary text-foreground"
                    : "border-transparent text-muted-foreground hover:bg-accent",
                )}
              >
                <CalendarDate size={14} />
                Gantt
              </button>
              {onSelectIssues && (
                <button
                  type="button"
                  onClick={onSelectIssues}
                  className={cn(
                    "flex w-full items-center justify-center gap-1 whitespace-nowrap rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                    activeView === "issues"
                      ? "border-border bg-secondary text-foreground"
                      : "border-transparent text-muted-foreground hover:bg-accent",
                  )}
                >
                  {t("tasks:issues.tab")}
                </button>
              )}
            </div>
          </div>

          <div className="space-y-1">
            <p className="px-1 text-[11px] font-medium tracking-wide text-muted-foreground uppercase">
              Projects
            </p>
            <div className="max-h-56 space-y-0.5 overflow-y-auto">
              {(projects ?? []).map((project) => {
                const Icon =
                  icons[project.icon as keyof typeof icons] || icons.Layout;
                const isCurrentProject = project.id === projectId;

                return (
                  <button
                    key={project.id}
                    type="button"
                    onClick={() => onSelectProject(project.id)}
                    className={cn(
                      "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors",
                      isCurrentProject
                        ? "bg-accent text-foreground"
                        : "text-muted-foreground hover:bg-accent hover:text-foreground",
                    )}
                  >
                    <Icon className="size-3.5" />
                    <span className="flex-1 truncate">{project.name}</span>
                    {isCurrentProject && <CheckRead size={14} />}
                  </button>
                );
              })}
            </div>
          </div>

          <button
            type="button"
            onClick={onAddProject}
            className="flex w-full items-center gap-2 rounded-md border border-border px-2 py-1.5 text-left text-sm text-foreground transition-colors hover:bg-accent"
          >
            <AddCircle size={14} />
            Add project
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
