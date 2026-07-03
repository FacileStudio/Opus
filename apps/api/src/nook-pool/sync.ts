import { PoolClient } from "@facile/pool";
import { createId } from "@paralleldrive/cuid2";
import { and, asc, eq, isNull, lt, max } from "drizzle-orm";
import db from "../database";
import {
  columnTable,
  nookPoolConfigTable,
  nookPoolOutboxTable,
  nookPoolProcessedEventTable,
  projectTable,
  taskTable,
  userTable,
} from "../database/schema";
import { publishEvent } from "../events";
import { DEFAULT_PROJECT_COLUMNS } from "../project/controllers/create-project";

const POOL_EVENT_VERSION = 1;

const POOL_EVENTS = {
  emit: [
    "project.created",
    "project.updated",
    "project.deleted",
    "task.created",
    "task.updated",
    "task.deleted",
  ],
  listen: [
    "project.created",
    "project.updated",
    "project.deleted",
    "task.created",
    "task.updated",
    "task.deleted",
  ],
};

export const ALL_POOL_EVENTS = [
  "task.created",
  "task.updated",
  "task.deleted",
  "project.created",
  "project.updated",
  "project.deleted",
] as const;

export type PoolEventToggle = {
  event: string;
  enabled: boolean;
};

export async function isPoolEventEnabled(
  workspaceId: string,
  event: string,
): Promise<boolean> {
  const [config] = await db
    .select({ poolEvents: nookPoolConfigTable.poolEvents })
    .from(nookPoolConfigTable)
    .where(eq(nookPoolConfigTable.workspaceId, workspaceId))
    .limit(1);

  if (!config || !config.poolEvents) return true;

  try {
    const toggles = JSON.parse(config.poolEvents) as Record<string, boolean>;
    return toggles[event] !== false;
  } catch {
    return true;
  }
}

const ICON_TYPO_MAP: Record<string, string> = {
  Pallete2: "Palette2",
  Siderbar: "Sidebar",
  Magnifer: "Magnifier",
  MagniferBug: "MagnifierBug",
  MagniferZoomIn: "MagnifierZoomIn",
  MagniferZoomOut: "MagnifierZoomOut",
  MinimalisticMagnifer: "MinimalisticMagnifier",
  MinimalisticMagniferBug: "MinimalisticMagnifierBug",
  MinimalisticMagniferZoomIn: "MinimalisticMagnifierZoomIn",
  MinimalisticMagniferZoomOut: "MinimalisticMagnifierZoomOut",
  RoundedMagnifer: "RoundedMagnifier",
  RoundedMagniferBug: "RoundedMagnifierBug",
  RoundedMagniferZoomIn: "RoundedMagnifierZoomIn",
  RoundedMagniferZoomOut: "RoundedMagnifierZoomOut",
  Condicioner: "Conditioner",
  Condicioner2: "Conditioner2",
  ColourTuneing: "ColourTuning",
  MaskHapply: "MaskHappy",
  SpedometerLow: "SpeedometerLow",
  SpedometerMax: "SpeedometerMax",
  SpedometerMiddle: "SpeedometerMiddle",
  CardRecive: "CardReceive",
  ReciveSquare: "ReceiveSquare",
  ReciveTwiceSquare: "ReceiveTwiceSquare",
  PlaaylistMinimalistic: "PlaylistMinimalistic",
};

function normalizeIcon(icon: string | null | undefined): string {
  if (!icon) return "Layout";
  return ICON_TYPO_MAP[icon] ?? icon;
}

const activeClients = new Map<string, PoolClient>();

export async function connectPoolForWorkspace(
  workspaceId: string,
  instanceUrl: string,
  secret: string,
): Promise<void> {
  disconnectPoolForWorkspace(workspaceId);

  const pool = new PoolClient({
    config: {
      app: "Opus",
      instance: instanceUrl,
      secret,
      instance_id: workspaceId,
      events: POOL_EVENTS,
    },
    onConnect: () => {
      console.log(`Pool connected for workspace ${workspaceId}`);
    },
    onDisconnect: () => {
      console.log(`Pool disconnected for workspace ${workspaceId}`);
    },
    onError: (error) => {
      console.error(`Pool error for workspace ${workspaceId}:`, error.message);
    },
    reconnect: true,
  });

  setupListeners(pool, workspaceId);

  await pool.connect();
  activeClients.set(workspaceId, pool);
}

export function disconnectPoolForWorkspace(workspaceId: string): void {
  const client = activeClients.get(workspaceId);
  if (client) {
    client.disconnect();
    activeClients.delete(workspaceId);
  }
}

export function getPoolClient(workspaceId: string): PoolClient | undefined {
  return activeClients.get(workspaceId);
}

let outboxTimer: ReturnType<typeof setInterval> | null = null;
const lastConnectAttempt = new Map<string, number>();

export async function initPoolSync(): Promise<void> {
  const envUrl = process.env.NOOK_POOL_URL;
  const envSecret = process.env.NOOK_POOL_SECRET;

  if (envUrl && envSecret) {
    console.log(
      "Pool: NOOK_POOL_URL and NOOK_POOL_SECRET set, available as defaults",
    );
  }

  const configs = await db
    .select()
    .from(nookPoolConfigTable)
    .where(eq(nookPoolConfigTable.enabled, true));

  for (const config of configs) {
    const url = config.instanceUrl || envUrl;
    const secret = config.secret || envSecret;

    if (!url || !secret) continue;

    try {
      await connectPoolForWorkspace(config.workspaceId, url, secret);
    } catch (error) {
      console.error(
        `Failed to init Pool for workspace ${config.workspaceId}:`,
        error,
      );
    }
  }

  await db
    .delete(nookPoolProcessedEventTable)
    .where(
      lt(
        nookPoolProcessedEventTable.processedAt,
        new Date(Date.now() - 35 * 24 * 60 * 60 * 1000),
      ),
    )
    .catch((err) => console.error("Pool: ledger prune failed:", err));

  if (!outboxTimer) {
    outboxTimer = setInterval(() => {
      drainOutbox().catch((err) =>
        console.error("Pool: outbox drain failed:", err),
      );
    }, 2000);
  }
}

export async function shutdownPoolSync(): Promise<void> {
  if (outboxTimer) {
    clearInterval(outboxTimer);
    outboxTimer = null;
  }
  for (const [workspaceId] of activeClients) {
    disconnectPoolForWorkspace(workspaceId);
  }
}

async function enqueueOutbox(
  workspaceId: string,
  channel: string,
  envelope: Record<string, unknown>,
): Promise<void> {
  await db.insert(nookPoolOutboxTable).values({
    workspaceId,
    channel,
    payload: JSON.stringify(envelope),
  });
}

const OUTBOX_BATCH = 100;
const CONNECT_RETRY_MS = 30_000;

async function drainOutbox(): Promise<void> {
  const configs = await db
    .select({
      workspaceId: nookPoolConfigTable.workspaceId,
      instanceUrl: nookPoolConfigTable.instanceUrl,
      secret: nookPoolConfigTable.secret,
    })
    .from(nookPoolConfigTable)
    .where(eq(nookPoolConfigTable.enabled, true));

  for (const config of configs) {
    let client = activeClients.get(config.workspaceId);

    if (!client?.isConnected) {
      const last = lastConnectAttempt.get(config.workspaceId) ?? 0;
      const url = config.instanceUrl || process.env.NOOK_POOL_URL;
      const secret = config.secret || process.env.NOOK_POOL_SECRET;
      if (url && secret && Date.now() - last > CONNECT_RETRY_MS) {
        lastConnectAttempt.set(config.workspaceId, Date.now());
        try {
          await connectPoolForWorkspace(config.workspaceId, url, secret);
          client = activeClients.get(config.workspaceId);
        } catch {
          continue;
        }
      }
      if (!client?.isConnected) continue;
    }
    lastConnectAttempt.set(config.workspaceId, Date.now());

    const rows = await db
      .select()
      .from(nookPoolOutboxTable)
      .where(eq(nookPoolOutboxTable.workspaceId, config.workspaceId))
      .orderBy(asc(nookPoolOutboxTable.seq))
      .limit(OUTBOX_BATCH);

    for (const row of rows) {
      try {
        if (!client.emitNow(row.channel, JSON.parse(row.payload))) {
          break;
        }
        await db
          .delete(nookPoolOutboxTable)
          .where(eq(nookPoolOutboxTable.id, row.id));
      } catch (error) {
        await db
          .update(nookPoolOutboxTable)
          .set({
            attempts: row.attempts + 1,
            lastError: error instanceof Error ? error.message : String(error),
          })
          .where(eq(nookPoolOutboxTable.id, row.id));
        break;
      }
    }
  }
}

async function isPoolEnabled(workspaceId: string): Promise<boolean> {
  if (activeClients.has(workspaceId)) return true;
  const [config] = await db
    .select({ enabled: nookPoolConfigTable.enabled })
    .from(nookPoolConfigTable)
    .where(eq(nookPoolConfigTable.workspaceId, workspaceId))
    .limit(1);
  return config?.enabled ?? false;
}

async function alreadyProcessed(
  workspaceId: string,
  key: string | undefined,
): Promise<boolean> {
  if (!key) return false;
  const [row] = await db
    .select({ key: nookPoolProcessedEventTable.idempotencyKey })
    .from(nookPoolProcessedEventTable)
    .where(
      and(
        eq(nookPoolProcessedEventTable.workspaceId, workspaceId),
        eq(nookPoolProcessedEventTable.idempotencyKey, key),
      ),
    )
    .limit(1);
  return !!row;
}

async function markProcessed(
  workspaceId: string,
  key: string | undefined,
): Promise<void> {
  if (!key) return;
  await db
    .insert(nookPoolProcessedEventTable)
    .values({ workspaceId, idempotencyKey: key })
    .onConflictDoNothing();
}

function createSyncQueue() {
  const queue: Array<() => Promise<void>> = [];
  let processing = false;

  async function drain() {
    if (processing) return;
    processing = true;
    let fn = queue.shift();
    while (fn) {
      try {
        await fn();
      } catch (error) {
        console.error("Pool sync queue error:", error);
      }
      fn = queue.shift();
    }
    processing = false;
  }

  return (fn: () => Promise<void>) => {
    queue.push(fn);
    drain();
  };
}

type Enqueue = (fn: () => Promise<void>) => void;

function setupListeners(pool: PoolClient, workspaceId: string): void {
  const enqueue = createSyncQueue();

  pool.listen("project.created", (payload: unknown) => {
    enqueue(() => handleIncomingProject("created", workspaceId, payload));
  });

  pool.listen("project.updated", (payload: unknown) => {
    enqueue(() => handleIncomingProject("updated", workspaceId, payload));
  });

  pool.listen("project.deleted", (payload: unknown) => {
    enqueue(() => handleIncomingProject("deleted", workspaceId, payload));
  });

  pool.listen("task.created", (payload: unknown) => {
    enqueue(() => handleIncomingTask("created", workspaceId, payload, enqueue));
  });

  pool.listen("task.updated", (payload: unknown) => {
    enqueue(() => handleIncomingTask("updated", workspaceId, payload, enqueue));
  });

  pool.listen("task.deleted", (payload: unknown) => {
    enqueue(() => handleIncomingTask("deleted", workspaceId, payload, enqueue));
  });
}

type IncomingProjectPayload = {
  facile_id: string;
  idempotency_key?: string;
  payload: {
    facile_id: string;
    name: string;
    description?: string;
    icon?: string | null;
  };
};

type IncomingTaskPayload = {
  facile_id: string;
  idempotency_key?: string;
  payload: {
    facile_id: string;
    project_facile_id: string;
    name: string;
    status?: string;
    actor_email?: string;
  };
};

async function handleIncomingProject(
  action: string,
  workspaceId: string,
  raw: unknown,
): Promise<void> {
  const data = raw as IncomingProjectPayload;
  const facileId = data.facile_id || data.payload?.facile_id;
  if (!facileId) return;
  if (await alreadyProcessed(workspaceId, data.idempotency_key)) return;

  if (action === "created") {
    const [existing] = await db
      .select({ id: projectTable.id })
      .from(projectTable)
      .where(
        and(
          eq(projectTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (existing) {
      await markProcessed(workspaceId, data.idempotency_key);
      return;
    }

    const slug =
      data.payload.name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-|-$/g, "") || "project";

    await db.transaction(async (tx) => {
      const [created] = await tx
        .insert(projectTable)
        .values({
          workspaceId,
          name: data.payload.name,
          description: data.payload.description || null,
          slug,
          facileId,
          icon: normalizeIcon(data.payload.icon),
        })
        .returning({ id: projectTable.id });

      if (created) {
        for (const col of DEFAULT_PROJECT_COLUMNS) {
          await tx.insert(columnTable).values({
            projectId: created.id,
            name: col.name,
            slug: col.slug,
            position: col.position,
            isFinal: col.isFinal,
          });
        }
      }
    });
    await markProcessed(workspaceId, data.idempotency_key);
  } else if (action === "updated") {
    const [existing] = await db
      .select({ id: projectTable.id })
      .from(projectTable)
      .where(
        and(
          eq(projectTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (!existing) {
      await handleIncomingProject("created", workspaceId, raw);
      return;
    }

    const updateData: Record<string, unknown> = {};
    if (data.payload.name) updateData.name = data.payload.name;
    if (data.payload.description !== undefined)
      updateData.description = data.payload.description;
    if (data.payload.icon !== undefined)
      updateData.icon = normalizeIcon(data.payload.icon);

    if (Object.keys(updateData).length > 0) {
      await db
        .update(projectTable)
        .set(updateData)
        .where(eq(projectTable.id, existing.id));
    }
    await markProcessed(workspaceId, data.idempotency_key);
  } else if (action === "deleted") {
    await db
      .delete(projectTable)
      .where(
        and(
          eq(projectTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      );
    await markProcessed(workspaceId, data.idempotency_key);
  }
}

async function resolveActorByEmail(
  email: string | undefined,
): Promise<string | null> {
  if (!email) return null;
  const [actor] = await db
    .select({ id: userTable.id })
    .from(userTable)
    .where(eq(userTable.email, email))
    .limit(1);
  return actor?.id ?? null;
}

const MAX_TASK_PARENT_RETRIES = 5;

async function handleIncomingTask(
  action: string,
  workspaceId: string,
  raw: unknown,
  enqueue: Enqueue,
  attempt = 0,
): Promise<void> {
  const data = raw as IncomingTaskPayload;
  const facileId = data.facile_id || data.payload?.facile_id;
  if (!facileId) return;
  if (await alreadyProcessed(workspaceId, data.idempotency_key)) return;

  if (action === "created") {
    const [existing] = await db
      .select({ id: taskTable.id, projectId: taskTable.projectId })
      .from(taskTable)
      .innerJoin(projectTable, eq(taskTable.projectId, projectTable.id))
      .where(
        and(
          eq(taskTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (existing) {
      const updateData: Record<string, unknown> = {};
      if (data.payload.name) updateData.title = data.payload.name;
      if (data.payload.status) {
        updateData.status = data.payload.status;
        const column = await db.query.columnTable.findFirst({
          where: and(
            eq(columnTable.projectId, existing.projectId),
            eq(columnTable.slug, data.payload.status),
          ),
        });
        if (column) updateData.columnId = column.id;
      }
      const actorId = await resolveActorByEmail(data.payload.actor_email);
      if (actorId) updateData.actorId = actorId;
      if (Object.keys(updateData).length > 0) {
        await db
          .update(taskTable)
          .set(updateData)
          .where(eq(taskTable.id, existing.id));

        await publishEvent("task.updated", {
          taskId: existing.id,
          projectId: existing.projectId,
          userId: "",
        });
      }
      await markProcessed(workspaceId, data.idempotency_key);
      return;
    }

    const projectFacileId = data.payload.project_facile_id;
    if (!projectFacileId) return;

    const [project] = await db
      .select({ id: projectTable.id })
      .from(projectTable)
      .where(
        and(
          eq(projectTable.facileId, projectFacileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (!project) {
      if (attempt < MAX_TASK_PARENT_RETRIES) {
        setTimeout(() => {
          enqueue(() =>
            handleIncomingTask(action, workspaceId, raw, enqueue, attempt + 1),
          );
        }, 1000);
      }
      return;
    }

    const status = data.payload.status || "to-do";

    const column = await db.query.columnTable.findFirst({
      where: and(
        eq(columnTable.projectId, project.id),
        eq(columnTable.slug, status),
      ),
    });

    const [maxResult] = await db
      .select({ maxNumber: max(taskTable.number) })
      .from(taskTable)
      .where(eq(taskTable.projectId, project.id));
    const nextNumber = (maxResult?.maxNumber ?? 0) + 1;

    const actorId = await resolveActorByEmail(data.payload.actor_email);

    const [created] = await db
      .insert(taskTable)
      .values({
        projectId: project.id,
        title: data.payload.name,
        status,
        columnId: column?.id ?? null,
        facileId,
        number: nextNumber,
        actorId,
      })
      .returning({ id: taskTable.id });

    if (created) {
      await publishEvent("task.created", {
        taskId: created.id,
        projectId: project.id,
        userId: "",
      });
    }
    await markProcessed(workspaceId, data.idempotency_key);
  } else if (action === "updated") {
    const [task] = await db
      .select({ id: taskTable.id, projectId: taskTable.projectId })
      .from(taskTable)
      .innerJoin(projectTable, eq(taskTable.projectId, projectTable.id))
      .where(
        and(
          eq(taskTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (!task) {
      await handleIncomingTask("created", workspaceId, raw, enqueue, attempt);
      return;
    }

    const updateData: Record<string, unknown> = {};
    if (data.payload.name) updateData.title = data.payload.name;
    if (data.payload.status) {
      updateData.status = data.payload.status;

      const column = await db.query.columnTable.findFirst({
        where: and(
          eq(columnTable.projectId, task.projectId),
          eq(columnTable.slug, data.payload.status),
        ),
      });
      if (column) updateData.columnId = column.id;
    }
    const actorId = await resolveActorByEmail(data.payload.actor_email);
    if (actorId) updateData.actorId = actorId;

    if (Object.keys(updateData).length > 0) {
      await db
        .update(taskTable)
        .set(updateData)
        .where(eq(taskTable.id, task.id));

      await publishEvent("task.updated", {
        taskId: task.id,
        projectId: task.projectId,
        userId: "",
      });
    }
    await markProcessed(workspaceId, data.idempotency_key);
  } else if (action === "deleted") {
    const [task] = await db
      .select({ id: taskTable.id })
      .from(taskTable)
      .innerJoin(projectTable, eq(taskTable.projectId, projectTable.id))
      .where(
        and(
          eq(taskTable.facileId, facileId),
          eq(projectTable.workspaceId, workspaceId),
        ),
      )
      .limit(1);

    if (task) {
      await db.delete(taskTable).where(eq(taskTable.id, task.id));
    }
    await markProcessed(workspaceId, data.idempotency_key);
  }
}

export async function emitProjectEvent(
  workspaceId: string,
  action: "created" | "updated" | "deleted",
  project: {
    facileId: string | null;
    name: string;
    description?: string | null;
    icon?: string | null;
  },
): Promise<void> {
  if (!(await isPoolEnabled(workspaceId))) return;

  const facileId = project.facileId;
  if (!facileId) return;

  const eventName = `project.${action}`;
  if (!(await isPoolEventEnabled(workspaceId, eventName))) return;

  await enqueueOutbox(workspaceId, eventName, {
    version: POOL_EVENT_VERSION,
    app: "Opus",
    object: "project",
    action,
    facile_id: facileId,
    payload: {
      facile_id: facileId,
      name: project.name,
      description: project.description || "",
      icon: normalizeIcon(project.icon) || null,
    },
    timestamp: new Date().toISOString(),
    idempotency_key: createId(),
  });
}

export async function emitTaskEvent(
  workspaceId: string,
  action: "created" | "updated" | "deleted",
  task: {
    facileId: string | null;
    title: string;
    status: string;
    projectId: string;
    actorId?: string | null;
  },
  projectFacileId: string | null,
): Promise<void> {
  if (!(await isPoolEnabled(workspaceId))) return;

  const facileId = task.facileId;
  if (!facileId) return;

  const eventName = `task.${action}`;
  if (!(await isPoolEventEnabled(workspaceId, eventName))) return;

  let actorEmail = "";
  if (task.actorId) {
    const [actor] = await db
      .select({ email: userTable.email })
      .from(userTable)
      .where(eq(userTable.id, task.actorId))
      .limit(1);
    actorEmail = actor?.email ?? "";
  }

  await enqueueOutbox(workspaceId, eventName, {
    version: POOL_EVENT_VERSION,
    app: "Opus",
    object: "task",
    action,
    facile_id: facileId,
    payload: {
      facile_id: facileId,
      project_facile_id: projectFacileId || "",
      name: task.title,
      status: task.status,
      actor_email: actorEmail,
    },
    timestamp: new Date().toISOString(),
    idempotency_key: createId(),
  });
}

export async function performInitialSync(
  workspaceId: string,
): Promise<{ projects_synced: number; tasks_synced: number }> {
  const client = activeClients.get(workspaceId);
  if (!client?.isConnected) {
    throw new Error("Pool is not connected for this workspace");
  }

  const unflaggedProjects = await db
    .select({ id: projectTable.id })
    .from(projectTable)
    .where(
      and(
        eq(projectTable.workspaceId, workspaceId),
        isNull(projectTable.facileId),
      ),
    );

  for (const p of unflaggedProjects) {
    await db
      .update(projectTable)
      .set({ facileId: createId() })
      .where(eq(projectTable.id, p.id));
  }

  const allProjects = await db
    .select({
      id: projectTable.id,
      name: projectTable.name,
      description: projectTable.description,
      facileId: projectTable.facileId,
    })
    .from(projectTable)
    .where(eq(projectTable.workspaceId, workspaceId));

  for (let i = 0; i < allProjects.length; i++) {
    const p = allProjects[i];
    if (!p.facileId) continue;
    emitProjectEvent(workspaceId, "created", p);
    if ((i + 1) % 50 === 0 && i < allProjects.length - 1) {
      await new Promise((r) => setTimeout(r, 100));
    }
  }

  if (allProjects.length > 0) {
    await new Promise((r) => setTimeout(r, 2000));
  }

  const unflaggedTasks = await db
    .select({ id: taskTable.id })
    .from(taskTable)
    .innerJoin(projectTable, eq(taskTable.projectId, projectTable.id))
    .where(
      and(
        eq(projectTable.workspaceId, workspaceId),
        isNull(taskTable.facileId),
      ),
    );

  for (const t of unflaggedTasks) {
    await db
      .update(taskTable)
      .set({ facileId: createId() })
      .where(eq(taskTable.id, t.id));
  }

  const allTasks = await db
    .select({
      id: taskTable.id,
      title: taskTable.title,
      status: taskTable.status,
      projectId: taskTable.projectId,
      facileId: taskTable.facileId,
      projectFacileId: projectTable.facileId,
    })
    .from(taskTable)
    .innerJoin(projectTable, eq(taskTable.projectId, projectTable.id))
    .where(eq(projectTable.workspaceId, workspaceId));

  for (let i = 0; i < allTasks.length; i++) {
    const t = allTasks[i];
    if (!t.facileId) continue;
    emitTaskEvent(
      workspaceId,
      "created",
      {
        facileId: t.facileId,
        title: t.title,
        status: t.status,
        projectId: t.projectId,
      },
      t.projectFacileId,
    );
    if ((i + 1) % 50 === 0 && i < allTasks.length - 1) {
      await new Promise((r) => setTimeout(r, 100));
    }
  }

  return { projects_synced: allProjects.length, tasks_synced: allTasks.length };
}
