import { client } from "@opus/libs";
import type { ProjectWithTasks } from "@/types/project";
import type Task from "@/types/task";

async function getSharedProject(token: string): Promise<ProjectWithTasks> {
  const response = await client["shared-project"][":token"].$get({
    param: { token },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  const { columns, tasks, ...project } = await response.json();

  const toTask = (task: (typeof tasks)[number]): Task => ({
    ...task,
    projectId: project.id,
    position: null,
  });

  return {
    ...project,
    columns: columns.map((column) => ({
      id: column.id,
      slug: column.slug,
      name: column.name,
      isFinal: column.isFinal,
      tasks: tasks.filter((task) => task.columnId === column.id).map(toTask),
    })),
    archivedTasks: [],
    plannedTasks: [],
  };
}

export default getSharedProject;
