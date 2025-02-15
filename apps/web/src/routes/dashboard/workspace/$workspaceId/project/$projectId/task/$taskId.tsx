import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/dashboard/workspace/$workspaceId/project/$projectId/task/$taskId",
)({
  component: TaskEditPage,
});

function TaskEditPage() {
  return <div>Task edit</div>;
}
