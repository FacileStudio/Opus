import descriptions from "@/constants/descriptions";
import { priorities } from "@/constants/priorities-and-statuses";
import taskTitles from "@/constants/task-titles";
import generateRandomDueDate from "./generate-random-due-date";

function generateTask({
  projectId,
  workspaceId,
  status,
}: {
  projectId: string;
  workspaceId: string;
  status: string;
}) {
  const id = Math.random().toString(36).substr(2, 9);

  return {
    id,
    title: taskTitles[Math.floor(Math.random() * taskTitles.length)],
    description: descriptions[Math.floor(Math.random() * descriptions.length)],
    assigneeId: Math.floor(Math.random() * 5) + 1, // Random assignee 1-5
    priority: priorities[Math.floor(Math.random() * priorities.length)],
    dueDate: Math.random() > 0.3 ? generateRandomDueDate() : null, // 70% chance of having a due date
    status,
    projectId,
    workspaceId,
  };
}

export default generateTask;
