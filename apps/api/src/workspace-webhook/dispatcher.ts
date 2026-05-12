import { createHmac } from "node:crypto";
import { and, eq } from "drizzle-orm";
import db from "../database";
import {
  projectTable,
  taskTable,
  userTable,
  workspaceWebhookTable,
} from "../database/schema";
import { subscribeToEvent } from "../events";

const OUTBOUND_TIMEOUT_MS = 10_000;

type EventData = Record<string, unknown>;

function buildTaskUrl(workspaceId: string, projectId: string, taskId: string) {
  const clientUrl = process.env.OPUS_CLIENT_URL || "http://localhost:5173";
  return `${clientUrl}/dashboard/workspace/${workspaceId}/project/${projectId}/task/${taskId}`;
}

async function resolveContext(data: EventData) {
  const projectId = typeof data.projectId === "string" ? data.projectId : null;
  const taskId = typeof data.taskId === "string" ? data.taskId : null;
  const currentUserId =
    typeof data.currentUserId === "string"
      ? data.currentUserId
      : typeof data.userId === "string"
        ? data.userId
        : null;

  let workspaceId: string | null =
    typeof data.workspaceId === "string" ? data.workspaceId : null;
  let projectName: string | null = null;

  if (projectId) {
    const [project] = await db
      .select({
        workspaceId: projectTable.workspaceId,
        name: projectTable.name,
      })
      .from(projectTable)
      .where(eq(projectTable.id, projectId))
      .limit(1);
    if (project) {
      workspaceId = project.workspaceId;
      projectName = project.name;
    }
  }

  if (!workspaceId) return null;

  let taskNumber: number | null = null;
  let taskTitle: string | null =
    typeof data.title === "string" ? data.title : null;
  let taskStatus: string | null =
    typeof data.status === "string" ? data.status : null;
  let taskPriority: string | null =
    typeof data.priority === "string" ? data.priority : null;

  if (taskId) {
    const [task] = await db
      .select({
        number: taskTable.number,
        title: taskTable.title,
        status: taskTable.status,
        priority: taskTable.priority,
      })
      .from(taskTable)
      .where(eq(taskTable.id, taskId))
      .limit(1);
    if (task) {
      taskNumber = task.number;
      taskTitle = taskTitle ?? task.title;
      taskStatus = taskStatus ?? task.status;
      taskPriority = taskPriority ?? task.priority;
    }
  }

  const actorId: string | null = currentUserId;
  let actorName: string | null = null;

  if (actorId) {
    const [user] = await db
      .select({ name: userTable.name })
      .from(userTable)
      .where(eq(userTable.id, actorId))
      .limit(1);
    actorName = user?.name ?? null;
  }

  return {
    workspaceId,
    project:
      projectId && projectName
        ? { id: projectId, name: projectName, workspaceId }
        : null,
    task: taskId
      ? {
          id: taskId,
          number: taskNumber,
          title: taskTitle,
          status: taskStatus,
          priority: taskPriority,
          url: projectId ? buildTaskUrl(workspaceId, projectId, taskId) : null,
        }
      : null,
    actor: actorId ? { id: actorId, name: actorName } : null,
  };
}

function buildEventData(eventType: string, data: EventData): EventData {
  const filtered: EventData = {};

  if (eventType === "task.status_changed") {
    if (data.oldStatus) filtered.oldStatus = data.oldStatus;
    if (data.newStatus) filtered.newStatus = data.newStatus;
  }
  if (eventType === "task.priority_changed") {
    if (data.oldPriority) filtered.oldPriority = data.oldPriority;
    if (data.newPriority) filtered.newPriority = data.newPriority;
  }
  if (eventType === "task.title_changed") {
    if (data.oldTitle) filtered.oldTitle = data.oldTitle;
    if (data.newTitle) filtered.newTitle = data.newTitle;
  }
  if (eventType === "task.description_changed") {
    if (data.oldDescription) filtered.oldDescription = data.oldDescription;
    if (data.newDescription) filtered.newDescription = data.newDescription;
  }
  if (
    eventType === "task.assignee_changed" ||
    eventType === "task.unassigned"
  ) {
    if (data.oldAssignee) filtered.oldAssignee = data.oldAssignee;
    if (data.newAssignee) filtered.newAssignee = data.newAssignee;
  }
  if (eventType === "task.due_date_changed") {
    if (data.oldDueDate) filtered.oldDueDate = data.oldDueDate;
    if (data.newDueDate) filtered.newDueDate = data.newDueDate;
  }
  if (eventType === "comment.created" || eventType === "task.comment_created") {
    if (data.content) filtered.comment = data.content;
  }

  return filtered;
}

async function deliverToWorkspaceWebhooks(eventType: string, data: EventData) {
  const context = await resolveContext(data);
  if (!context) return;

  const webhooks = await db
    .select()
    .from(workspaceWebhookTable)
    .where(
      and(
        eq(workspaceWebhookTable.workspaceId, context.workspaceId),
        eq(workspaceWebhookTable.enabled, true),
      ),
    );

  if (!webhooks.length) return;

  const payload = JSON.stringify({
    event: eventType,
    timestamp: new Date().toISOString(),
    project: context.project,
    task: context.task,
    actor: context.actor,
    data: buildEventData(eventType, data),
  });

  const deliveries = webhooks
    .filter((webhook) => {
      if (webhook.events === "*") return true;
      const allowed = webhook.events.split(",").map((e) => e.trim());
      return allowed.includes(eventType);
    })
    .map(async (webhook) => {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Opus-Event": eventType,
        "X-Opus-Webhook-Id": webhook.id,
      };

      if (webhook.secret) {
        headers["X-Opus-Signature"] = createHmac("sha256", webhook.secret)
          .update(payload)
          .digest("hex");
      }

      try {
        await fetch(webhook.url, {
          method: "POST",
          headers,
          body: payload,
          signal: AbortSignal.timeout(OUTBOUND_TIMEOUT_MS),
        });
      } catch (err) {
        console.error(
          `Workspace webhook delivery failed for ${webhook.id}:`,
          err instanceof Error ? err.message : err,
        );
      }
    });

  await Promise.allSettled(deliveries);
}

const TRACKED_EVENTS = [
  "task.created",
  "task.updated",
  "task.deleted",
  "task.status_changed",
  "task.assignee_changed",
  "task.priority_changed",
  "task.title_changed",
  "task.description_changed",
  "task.due_date_changed",
  "task.moved",
  "task.unassigned",
  "task.label_assigned",
  "task.label_unassigned",
  "task.comment_created",
  "comment.created",
  "comment.updated",
  "comment.deleted",
  "time-entry.created",
];

export function initializeWorkspaceWebhookDispatcher() {
  for (const eventType of TRACKED_EVENTS) {
    subscribeToEvent(eventType, async (data) => {
      try {
        await deliverToWorkspaceWebhooks(eventType, data as EventData);
      } catch (err) {
        console.error(
          `Workspace webhook dispatcher error for ${eventType}:`,
          err,
        );
      }
    });
  }
}
