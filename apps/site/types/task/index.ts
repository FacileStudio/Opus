type Task = {
  id: string;
  number: number;
  title: string;
  status: string;
  priority: string | null;
  dueDate: string | null;
  actorName?: string | null;
  actorImage?: string | null;
  description: string | null;
  columnId?: string | null;
  projectId: string;
  workspaceId: string;
  userId: string | null;
  position: number;
  createdAt: string;
  updatedAt: string;
};

export default Task;
