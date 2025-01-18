import { cn } from "@/lib/utils";
import useProjectStore from "@/store/project";
import { Layout } from "lucide-react";

type ProjectItemProps = {
  project: {
    projectId: string;
    projectName: string;
  };
};

function ProjectItem({ project }: ProjectItemProps) {
  const { project: selectedProject, setProject } = useProjectStore();
  const isSelected = project?.projectId === selectedProject?.projectId;

  return (
    <button
      type="button"
      key={project.projectId}
      onClick={() => setProject(project)}
      className={cn(
        "w-full text-left px-2 py-1.5 rounded-md flex items-center text-sm transition-all group",
        isSelected
          ? "bg-indigo-50 dark:bg-indigo-500/10"
          : "hover:bg-zinc-100 dark:hover:bg-zinc-800/50",
      )}
    >
      <Layout
        className={cn(
          "w-4 h-4 mr-2",
          isSelected
            ? "text-indigo-600 dark:text-indigo-400"
            : "text-zinc-400 dark:text-zinc-500 group-hover:text-zinc-500 dark:group-hover:text-zinc-400",
        )}
      />
      <span
        className={cn(
          "font-medium",
          isSelected
            ? "text-indigo-600 dark:text-indigo-400"
            : "text-zinc-600 dark:text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-zinc-300",
        )}
      >
        {project.projectName}
      </span>
    </button>
  );
}

export default ProjectItem;
