import type { Env, Hono } from "hono";

type Endpoint<
  Input extends object = object,
  Output = unknown,
  Status extends number = 200,
> = {
  input: Input;
  output: Output;
  outputFormat: "json";
  status: Status;
};

type Id = string;
type Timestamp = string;

export type LabelRecord = {
  id: Id;
  name: string;
  color: string;
  createdAt: Timestamp;
  taskId: Id | null;
  workspaceId: Id;
};

export type ColumnRecord = {
  id: Id;
  projectId: Id;
  name: string;
  slug: string;
  position: number;
  icon: string;
  color: string;
  isFinal: boolean;
  createdAt: Timestamp;
  updatedAt: Timestamp;
};

export type ImageUploadRecord = {
  key: string;
  uploadUrl: string;
  headers: Record<string, string>;
};

export type AssetRecord = {
  id: Id;
  url: string;
};

export type ExternalLinkRecord = {
  id: Id;
  taskId: Id;
  integrationId: Id;
  resourceType: string;
  externalId: string;
  url: string;
  title: string | null;
  metadata: Record<string, unknown> | null;
};

export type TaskRecord = {
  id: Id;
  projectId: Id;
  workspaceId: Id;
  columnId?: Id | null;
  number: number;
  position: number;
  title: string;
  description: string;
  status: string;
  priority: string;
  actorId?: Id | null;
  userId?: Id | null;
  actorName?: string;
  actorImage?: string;
  startDate?: Timestamp | null;
  dueDate?: Timestamp | null;
  createdAt: Timestamp;
  updatedAt: Timestamp;
  labels: LabelRecord[];
  subtaskTotal: number;
  subtaskDone: number;
};

export type ProjectStatistics = {
  completionPercentage: number;
  totalTasks: number;
  dueDate: Timestamp | null;
};

export type ProjectRecord = {
  id: Id;
  workspaceId: Id;
  slug: string;
  icon: string;
  name: string;
  description: string;
  isPublic: boolean;
  archivedAt: Timestamp | null;
  createdAt: Timestamp;
  statistics?: ProjectStatistics;
};

export type ProjectListRecord = ProjectRecord & {
  statistics: ProjectStatistics;
};

export type ActivityRecord = {
  id: Id;
  taskId: Id;
  userId: Id | null;
  type: string;
  content: string | null;
  eventData: Record<string, unknown> | null;
  createdAt: Timestamp;
  updatedAt: Timestamp;
};

export type TimeEntryRecord = {
  id: Id;
  taskId: Id;
  userId?: Id | null;
  description: string;
  startTime: Timestamp;
  endTime?: Timestamp | null;
  duration: number;
  createdAt: Timestamp;
  user?: { id: Id; name: string; image: string } | null;
};

export type TaskSummaryRecord = {
  id: Id;
  title: string;
  status: string;
  priority: string;
  number: number;
  projectId: Id;
  userId: Id | null;
  actorName: string | null;
};

export type TaskRelationRecord = {
  id: Id;
  sourceTaskId: Id;
  targetTaskId: Id;
  relationType: string;
  createdAt: Timestamp;
  sourceTask?: TaskSummaryRecord | null;
  targetTask?: TaskSummaryRecord | null;
};

export type ShareLinkRecord = {
  id: Id;
  projectId: Id;
  token: string;
  label: string;
  expiresAt: Timestamp | null;
  isActive: boolean;
  lastAccessedAt: Timestamp | null;
  accessCount: number;
  createdBy: Id;
  createdAt: Timestamp;
  updatedAt: Timestamp;
};

export type WorkflowRuleRecord = {
  id: Id;
  projectId: Id;
  integrationType: string;
  eventType: string;
  columnId: Id;
  createdAt: Timestamp;
  updatedAt: Timestamp;
};

export type SearchItem = {
  id: Id;
  type: string;
  title: string;
  description: string;
  content?: string;
  createdAt: Timestamp;
  relevanceScore: number;
  taskNumber?: number;
  priority?: string;
  status?: string;
  projectId?: Id;
  projectName?: string;
  projectSlug?: string;
  workspaceId?: Id;
  workspaceName?: string;
  userId?: Id;
  userName?: string;
};

export type SearchResponse = {
  results: SearchItem[];
  totalCount: number;
  searchQuery: string;
};

export type PublicProjectRecord = TasksApiResponse["data"];

export type SharedProjectRecord = {
  id: Id;
  name: string;
  slug: string;
  icon: string;
  description: string;
  isPublic: boolean;
  workspaceId: Id;
  columns: Array<Omit<ColumnRecord, "projectId" | "createdAt" | "updatedAt">>;
  tasks: Array<{
    id: Id;
    number: number;
    title: string;
    description: string;
    status: string;
    priority: string;
    columnId?: Id | null;
    dueDate?: Timestamp | null;
    createdAt: Timestamp;
  }>;
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

export type TasksApiResponse = {
  data: {
    id: Id;
    name: string;
    slug: string;
    icon: string;
    description: string;
    isPublic: boolean;
    workspaceId: Id;
    columns: Array<{
      id: Id;
      slug: string;
      name: string;
      isFinal: boolean;
      tasks: TaskRecord[];
    }>;
    archivedTasks: TaskRecord[];
    plannedTasks: TaskRecord[];
  };
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
  };
};

export type ConfigRecord = {
  ssoOnly: boolean;
  disableRegistration: boolean;
  disablePasswordRegistration: boolean;
  isDemoMode: boolean;
  hasSmtp: boolean;
  hasGithubSignIn: boolean;
  hasGoogleSignIn: boolean;
  hasDiscordSignIn: boolean;
  hasCustomOAuth: boolean;
  hasGuestAccess: boolean;
  customOAuthLogoutUrl: string | null;
};

export type ImportTasksResponse = {
  importedAt: Timestamp;
  results: { total: number; successful: number; failed: number };
};

export type AppSchema = {
  "/activity/comment": {
    $post: Endpoint<
      { json: { taskId: string; comment: string } },
      ActivityRecord
    >;
    $put: Endpoint<
      { json: { activityId: string; comment: string } },
      ActivityRecord
    >;
    $delete: Endpoint<{ json: { activityId: string } }>;
  };
  "/activity/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, ActivityRecord[]>;
  };
  "/column/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }, ColumnRecord[]>;
    $post: Endpoint<
      { param: { projectId: string }; json?: unknown },
      ColumnRecord
    >;
  };
  "/column/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, ColumnRecord>;
    $delete: Endpoint<{ param: { id: string } }, ColumnRecord>;
  };
  "/column/reorder/:projectId": {
    $put: Endpoint<{ param: { projectId: string }; json?: unknown }>;
  };
  "/config": {
    $get: Endpoint<object, ConfigRecord>;
  };
  "/external-link/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, ExternalLinkRecord[]>;
  };
  "/label": {
    $post: Endpoint<
      {
        json: {
          name: string;
          color: string;
          taskId?: string;
          workspaceId: string;
        };
      },
      LabelRecord
    >;
  };
  "/label/:id": {
    $put: Endpoint<
      {
        param: { id: string };
        json: { name?: string; color?: string };
      },
      LabelRecord
    >;
    $delete: Endpoint<{ param: { id: string } }, LabelRecord>;
  };
  "/label/workspace/:workspaceId": {
    $get: Endpoint<{ param: { workspaceId: string } }, LabelRecord[]>;
  };
  "/label/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, LabelRecord[]>;
  };
  "/notification": {
    $get: Endpoint<object, NotificationRecord[]>;
  };
  "/notification/clear-all": {
    $delete: Endpoint<object>;
  };
  "/notification/read-all": {
    $patch: Endpoint<object>;
  };
  "/notification/:id": {
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/notification/:id/read": {
    $patch: Endpoint<{ param: { id: string } }>;
  };
  "/notification/:id/unread": {
    $patch: Endpoint<{ param: { id: string } }>;
  };
  "/project": {
    $get: Endpoint<{ query: { workspaceId?: string } }, ProjectListRecord[]>;
    $post: Endpoint<
      {
        json: {
          name: string;
          slug: string;
          workspaceId: string;
          icon?: string;
        };
      },
      ProjectRecord
    >;
  };
  "/project/:id": {
    $get: Endpoint<
      { param: { id: string }; query?: { workspaceId?: string } },
      ProjectListRecord
    >;
    $put: Endpoint<
      {
        param: { id: string };
        json: {
          name?: string;
          icon?: string;
          slug?: string;
          description?: string;
          isPublic?: boolean;
        };
      },
      ProjectRecord
    >;
    $delete: Endpoint<{ param: { id: string } }, ProjectRecord>;
  };
  "/public-project/:id": {
    $get: Endpoint<{ param: { id: string } }, PublicProjectRecord>;
  };
  "/search": {
    $get: Endpoint<{ query?: Record<string, unknown> }, SearchResponse>;
  };
  "/share-link": {
    $get: Endpoint<{ query?: Record<string, unknown> }, ShareLinkRecord[]>;
    $post: Endpoint<{ json?: unknown }, ShareLinkRecord>;
  };
  "/share-link/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, ShareLinkRecord>;
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/shared-project/:token": {
    $get: Endpoint<{ param: { token: string } }, SharedProjectRecord>;
  };
  "/task-relation": {
    $post: Endpoint<{ json?: unknown }, TaskRelationRecord>;
  };
  "/task-relation/:id": {
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/task-relation/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, TaskRelationRecord[]>;
  };
  "/task/:projectId": {
    $post: Endpoint<
      {
        param: { projectId: string };
        json: {
          title: string;
          description: string;
          userId?: string;
          status: string;
          startDate?: string;
          dueDate?: string;
          priority: string;
        };
      },
      TaskRecord
    >;
  };
  "/task/:id": {
    $get: Endpoint<{ param: { id: string } }, TaskRecord>;
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
    $delete: Endpoint<{ param: { id: string } }, TaskRecord>;
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
    $put: Endpoint<
      {
        param: { id: string };
        json?: unknown;
        form?: unknown;
      },
      ImageUploadRecord
    >;
  };
  "/task/image-upload/:id/finalize": {
    $post: Endpoint<{ param: { id: string }; json?: unknown }, AssetRecord>;
  };
  "/task/export/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }>;
  };
  "/task/import/:projectId": {
    $post: Endpoint<
      { param: { projectId: string }; json?: unknown; form?: unknown },
      ImportTasksResponse
    >;
  };
  "/task/move/:id": {
    $put: Endpoint<
      { param: { id: string }; json?: unknown },
      { task: TaskRecord; sourceProjectId: Id; destinationProjectId: Id }
    >;
  };
  "/task/actor/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/task/description/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/task/due-date/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/task/priority/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/task/status/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/task/title/:id": {
    $put: Endpoint<{ param: { id: string }; json?: unknown }, TaskRecord>;
  };
  "/time-entry": {
    $post: Endpoint<
      {
        json: {
          taskId: string;
          description?: string;
          startTime: string;
        };
      },
      TimeEntryRecord
    >;
  };
  "/time-entry/:id": {
    $put: Endpoint<
      {
        param: { id: string };
        json: {
          startTime?: string;
          endTime?: string;
          description?: string;
        };
      },
      TimeEntryRecord
    >;
  };
  "/time-entry/task/:taskId": {
    $get: Endpoint<{ param: { taskId: string } }, TimeEntryRecord[]>;
  };
  "/workflow-rule/:id": {
    $delete: Endpoint<{ param: { id: string } }>;
  };
  "/workflow-rule/:projectId": {
    $get: Endpoint<{ param: { projectId: string } }, WorkflowRuleRecord[]>;
    $put: Endpoint<
      { param: { projectId: string }; json?: unknown },
      WorkflowRuleRecord[]
    >;
  };
};

export type AppType = Hono<Env, AppSchema, "/">;
