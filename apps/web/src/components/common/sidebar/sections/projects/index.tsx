import useGetProjects from "@/hooks/queries/project/use-get-projects";
import ProjectItem from "./project-item";

type ProjectsProps = {
  workspaceId: string;
};

function Projects({ workspaceId }: ProjectsProps) {
  const { data: projects } = useGetProjects({ workspaceId });

  return (
    <div className="ml-8 space-y-0.5 mt-1">
      {projects &&
        projects.length > 0 &&
        projects.map((project) => (
          <ProjectItem
            key={project.id}
            workspaceId={workspaceId}
            projectId={project?.id}
          />
        ))}
    </div>
  );
}

export default Projects;
