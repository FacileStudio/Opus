import createTask from "@/fetchers/task/create-task";
import { useMutation } from "@tanstack/react-query";

type CreateTaskInput = {
  title: string;
  description: string;
  assigneeId: string;
  projectId: string;
  status: string;
  dueDate: Date;
  number?: number;
  priority: "low" | "medium" | "high" | "urgent";
};

function useCreateTask() {
  return useMutation({
    mutationFn: ({
      title,
      description,
      assigneeId,
      projectId,
      status,
      dueDate,
      priority,
    }: CreateTaskInput) =>
      createTask(
        title,
        description,
        projectId,
        assigneeId,
        status,
        dueDate,
        priority,
      ),
  });
}

export default useCreateTask;
