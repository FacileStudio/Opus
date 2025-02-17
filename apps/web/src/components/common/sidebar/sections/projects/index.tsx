import useGetProjects from "@/hooks/queries/project/use-get-projects";
import { cn } from "@/lib/cn";
import useProjectStore from "@/store/project";
import { useUserPreferencesStore } from "@/store/user-preferences";
import type { Project } from "@/types/project";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { Layout, Plus, icons } from "lucide-react";
import { createElement, useState } from "react";
import CreateProjectModal from "./create-project-modal";

type ProjectsProps = {
  workspaceId: string;
};

function Projects({ workspaceId }: ProjectsProps) {
  const { project: currentProject, setProject } = useProjectStore();
  const { data: projects, isLoading } = useGetProjects({ workspaceId });
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false);
  const { isSidebarOpened } = useUserPreferencesStore();
  const navigate = useNavigate();
  const location = useLocation();

  const handleSelectProject = (selectedProject: Project) => {
    if (
      currentProject?.id === selectedProject.id &&
      location.pathname.includes("/board")
    )
      return;

    setProject(selectedProject);
    navigate({
      to: "/dashboard/workspace/$workspaceId/project/$projectId/board",
      params: {
        workspaceId,
        projectId: selectedProject.id,
      },
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-1 w-full">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="animate-pulse flex items-center px-4 py-1.5 rounded-md"
          >
            <div className="h-4 shrink-0 w-4 bg-zinc-200 dark:bg-zinc-800 rounded mr-2" />
            <div className="h-4 w-full bg-zinc-200 dark:bg-zinc-800 rounded" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div>
      <div
        className={cn(
          "flex items-center justify-between mb-2",
          isSidebarOpened && "px-3",
        )}
      >
        <h2 className="text-xs font-semibold text-zinc-500 dark:text-zinc-400 uppercase tracking-wider">
          {isSidebarOpened && "Projects"}
        </h2>
        <button
          type="button"
          onClick={() => setIsCreateProjectOpen(true)}
          className={cn(
            "rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors",
            !isSidebarOpened ? "p-2" : "p-1",
          )}
        >
          <Plus
            className={cn(
              "text-zinc-500 dark:text-zinc-400",
              !isSidebarOpened ? "w-5 h-5" : "w-4 h-4",
            )}
          />
        </button>
      </div>
      <div className="space-y-0.5">
        {projects && projects.length > 0 ? (
          projects.map((project) => (
            <button
              type="button"
              key={project.id}
              onClick={() => handleSelectProject(project)}
              className={cn(
                "w-full px-4 py-2 rounded-md flex items-center  text-sm transition-all group",
                !isSidebarOpened && "px-3",
                !isSidebarOpened && "justify-center px-2",
                location.pathname.includes("/board") &&
                  project.id === currentProject?.id
                  ? "bg-indigo-51 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400"
                  : "text-zinc-601 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800",
              )}
            >
              {createElement(
                icons[project.icon as keyof typeof icons] || Layout,
                {
                  className: cn(
                    "shrink-0",
                    !isSidebarOpened ? "w-6 h-6" : "w-4 h-4 mr-2",
                  ),
                },
              )}
              {isSidebarOpened && project.name}
            </button>
          ))
        ) : isSidebarOpened ? (
          <div className="px-3 py-4 flex flex-col items-center text-center">
            <div className="w-12 h-12 mb-3 rounded-full bg-indigo-50 dark:bg-indigo-500/10 flex items-center justify-center">
              <Layout className="w-6 h-6 text-indigo-500 dark:text-indigo-400" />
            </div>
            <p className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Create your first project
            </p>
            <p className="text-xs text-zinc-500 dark:text-zinc-400 mb-3">
              Start organizing your work with a new project
            </p>
            <button
              type="button"
              onClick={() => setIsCreateProjectOpen(true)}
              className="text-xs px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 dark:bg-indigo-500/10 dark:hover:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-md transition-colors font-medium inline-flex items-center gap-1.5"
            >
              <Plus className="w-3 h-3" />
              New Project
            </button>
          </div>
        ) : null}
      </div>
      <CreateProjectModal
        open={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
      />
    </div>
  );
}

export default Projects;
