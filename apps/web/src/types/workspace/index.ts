export type Workspace = {
  id: string;
  name: string;
  ownerId: string;
};

export type Column = {
  id: string;
  name: string;
  tasks: Task[];
};

export type Task = {
  id: string;
  title: string;
  description: string;
  assigneeId: number;
  priority: string;
  dueDate: string | null;
  status: string;
  projectId: string;
  workspaceId: string;
};
