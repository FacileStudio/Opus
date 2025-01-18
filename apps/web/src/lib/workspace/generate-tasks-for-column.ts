import generateTask from "./generate-task";

function generateTasksForColumn(
  count: number,
  columnId: string,
  projectId: string,
  workspaceId: string,
) {
  return Array(count)
    .fill(null)
    .map(() =>
      generateTask({
        projectId,
        workspaceId,
        status: columnId,
      }),
    );
}

export default generateTasksForColumn;
