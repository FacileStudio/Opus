export type Project = {
  id: string;
  name: string;
  description: string | null;
  workspaceId: string;
  columns?: Column[];
};

export type Column = {
  id: string;
  name: string;
  tasks: Task[];
};

export type Task = {
  id: string;
  createdAt: Date;
  description: string | null;
  projectId: string;
  assigneeId: string;
  title: string;
  status: string;
  dueDate: Date | null;
  priority: string | null;
  assigneeName: string;
};
