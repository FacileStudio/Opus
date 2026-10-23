import type { Env, Hono } from "hono";

type Endpoint<
  Input extends object = Record<string, never>,
  Output = unknown,
  Status extends number = 200,
> = {
  input: Input;
  output: Output;
  outputFormat: "json";
  status: Status;
};

export type NotificationRecord = {
  id: string;
  userId: string;
  title: string | null;
  content: string | null;
  type:
    | "info"
    | "task_created"
    | "workspace_created"
    | "task_status_changed"
    | "task_actor_changed"
    | "time_entry_created"
    | "due_date_reminder"
    | "task_overdue";
  eventData: Record<string, unknown> | null;
  isRead?: boolean;
  resourceId?: string;
  resourceType?: "task" | "workspace";
  createdAt: Date | string;
};

export type ProjectRecord = {
  id: string;
  name: string;
  slug: string;
  icon?: string | null;
  description?: string | null;
  isPublic?: boolean;
  workspaceId: string;
  createdAt?: string;
  updatedAt?: string;
  [key: string]: unknown;
};

export type MyTask = {
  id: string;
  projectId: string;
  projectSlug: string;
  projectName: string;
  projectIcon: string;
  number: number;
  title: string;
  status: string;
  priority: string;
  dueDate?: string | null;
  subtaskTotal: number;
  subtaskDone: number;
};

export type TasksApiResponse = {
  data: {
    id: string;
    name: string;
    slug: string;
    icon: string | null;
    description: string | null;
    isPublic: boolean;
    workspaceId: string;
    columns: Array<{
      id: string;
      name: string;
      position: number;
      color?: string | null;
      tasks?: unknown[];
      [key: string]: unknown;
    }>;
    archivedTasks: unknown[];
    plannedTasks: unknown[];
    [key: string]: unknown;
  };
  pagination?: {
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
  };
};

export type AppSchema = {
  "/activity/comment": {
    $post: Endpoint<{ json: { taskId: string; comment: string } }>;
    $put: Endpoint<{ json: { activityId: string; comment: string } }>;
    $delete: Endpoint<{ json: { activityId: string } }>;
  };
  "/activity/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }>;
  };
  "/column/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }>;
    $post: Endpoint<{ param: { projectId: string }; json?: unknown }>;
  };
  "/column/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/column/reorder/:projectId": {
    $put: Endpoint<{ param: { projectId: string }; json?: unknown }>;
  };
  "/config": {
    $get: Endpoint<Record<string, never>, Record<string, unknown>>;
  };
  "/external-link/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, unknown[]>;
  };
  "/label": {
    $post: Endpoint<{
      json: {
        name: string;
        color: string;
        taskId?: string;
        workspaceId: string;
      };
    }>;
  };
  "/label/:id": {
    $put: Endpoint<{
      param: { id: string };
      json: { name?: string; color?: string };
    }>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/label/workspace/:workspaceId": {
    $get: Endpoint<{ param: { workspaceId: string } }>;
  };
  "/label/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }>;
  };
  "/notification": {
    $get: Endpoint<Record<string, never>, NotificationRecord[]>;
  };
  "/notification/clear-all": {
    $delete: Endpoint<Record<string, never>>;
  };
  "/notification/read-all": {
    $patch: Endpoint<Record<string, never>>;
  };
  "/notification/:id/read": {
    $patch: Endpoint<{ param: { id: string } }>;
  };
  "/project": {
    $get: Endpoint<{ query?: { workspaceId?: string } }>;
    $post: Endpoint<{
      json: {
        name: string;
        slug: string;
        workspaceId: string;
        icon?: string;
      };
    }>;
  };
  "/project/:id": {
    $get: Endpoint<
      { param: { id: string }; query?: { workspaceId?: string } },
      ProjectRecord
    >;
    $put: Endpoint<{
      param: { id: string };
      json: {
        name?: string;
        icon?: string;
        slug?: string;
        description?: string;
        isPublic?: boolean;
      };
    }>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/public-project/:id": {
    $get: Endpoint<{ param: { id: string } }>;
  };
  "/search": {
    $get: Endpoint<{ query?: Record<string, unknown> }>;
  };
  "/share-link": {
    $get: Endpoint<{ query?: Record<string, unknown> }>;
    $post: Endpoint<{ json?: unknown }>;
  };
  "/share-link/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/shared-project/:token": {
    $get: Endpoint<{ param: { token: string } }>;
  };
  "/task-relation": {
    $post: Endpoint<{ json?: unknown }>;
  };
  "/task-relation/:id": {
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/task-relation/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }>;
  };
  "/task/:projectId": {
    $post: Endpoint<{
      param: { projectId: string };
      json: {
        title: string;
        description?: string;
        userId?: string;
        status?: string;
        startDate?: string;
        dueDate?: string;
        priority?: string;
      };
    }>;
  };
  "/task/me": {
    $get: Endpoint<
      { query: { workspaceId: string; includeDone?: string } },
      MyTask[]
    >;
  };
  "/task/:id": {
    $get: Endpoint<{ param: { id: string } }>;
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/task/tasks/:projectId": {
    $get: Endpoint<
      { param: { projectId: string }; query?: Record<string, unknown> },
      TasksApiResponse
    >;
  };
  "/task/bulk": {
    $patch: Endpoint<{ json?: unknown }>;
  };
  "/task/image-upload/:id": {
    $put: Endpoint<{
      param: { id: string };
      json?: unknown;
      form?: unknown;
    }>;
  };
  "/task/image-upload/:id/finalize": {
    $post: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/export/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }>;
  };
  "/task/import/:projectId": {
    $post: Endpoint<{ param: { projectId: string }; form?: unknown }>;
  };
  "/task/move/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/actor/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/description/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/due-date/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/priority/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/status/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/task/title/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }>;
  };
  "/time-entry": {
    $post: Endpoint<{
      json: {
        taskId: string;
        description?: string;
        startTime: string;
      };
    }>;
  };
  "/time-entry/:id": {
    $put: Endpoint<{
      param: { id: string };
      json: {
        startTime?: string;
        endTime?: string;
        description?: string;
      };
    }>;
  };
  "/time-entry/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }>;
  };
  "/workflow-rule/:id": {
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/workflow-rule/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }>;
    $put: Endpoint<{ param: { projectId: string }; json?: unknown }>;
  };
};

export type AppType = Hono<Env, AppSchema, "/">;
