import generateTasksForColumn from "./generate-tasks-for-column";

function generateProject({
  projectId = "1",
  workspaceId = "1",
  tasksPerColumn = 3,
}) {
  return {
    id: projectId,
    name: `Project ${projectId}`,
    workspaceId,
    columns: [
      {
        id: "todo",
        name: "To Do",
        tasks: generateTasksForColumn(
          tasksPerColumn,
          "todo",
          projectId,
          workspaceId,
        ),
      },
      {
        id: "in-progress",
        name: "In Progress",
        tasks: generateTasksForColumn(
          Math.ceil(tasksPerColumn / 2),
          "in-progress",
          projectId,
          workspaceId,
        ),
      },
      {
        id: "in-review",
        name: "In Review",
        tasks: generateTasksForColumn(
          Math.ceil(tasksPerColumn / 3),
          "in-review",
          projectId,
          workspaceId,
        ),
      },
      {
        id: "done",
        name: "Done",
        tasks: generateTasksForColumn(
          tasksPerColumn,
          "done",
          projectId,
          workspaceId,
        ),
      },
    ],
  };
}

export default generateProject;
