import { PoolClient } from "@facile/pool";
import { createId } from "@paralleldrive/cuid2";
import { and, eq, isNull, max } from "drizzle-orm";
import db from "../database";
import {
  columnTable,
  nookPoolConfigTable,
  projectTable,
  taskTable,
} from "../database/schema";
import { DEFAULT_PROJECT_COLUMNS } from "../project/controllers/create-project";

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
      app: "opus",
      instance: instanceUrl,
      secret,
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
}

export async function shutdownPoolSync(): Promise<void> {
  for (const [workspaceId] of activeClients) {
    disconnectPoolForWorkspace(workspaceId);
  }
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
    enqueue(() => handleIncomingTask("created", workspaceId, payload));
  });

  pool.listen("task.updated", (payload: unknown) => {
    enqueue(() => handleIncomingTask("updated", workspaceId, payload));
  });

  pool.listen("task.deleted", (payload: unknown) => {
    enqueue(() => handleIncomingTask("deleted", workspaceId, payload));
  });
}

type IncomingProjectPayload = {
  facile_id: string;
  payload: {
    facile_id: string;
    name: string;
    description?: string;
    icon?: string | null;
  };
};

type IncomingTaskPayload = {
  facile_id: string;
  payload: {
    facile_id: string;
    project_facile_id: string;
    name: string;
    status?: string;
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

  if (action === "created") {
    const [existing] = await db
      .select({ id: projectTable.id })
      .from(projectTable)
      .where(eq(projectTable.facileId, facileId))
      .limit(1);

    if (existing) return;

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
  } else if (action === "updated") {
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
        .where(eq(projectTable.facileId, facileId));
    }
  } else if (action === "deleted") {
    await db.delete(projectTable).where(eq(projectTable.facileId, facileId));
  }
}

async function handleIncomingTask(
  action: string,
  workspaceId: string,
  raw: unknown,
): Promise<void> {
  const data = raw as IncomingTaskPayload;
  const facileId = data.facile_id || data.payload?.facile_id;
  if (!facileId) return;

  if (action === "created") {
    const [existing] = await db
      .select({ id: taskTable.id })
      .from(taskTable)
      .where(eq(taskTable.facileId, facileId))
      .limit(1);

    if (existing) return;

    const projectFacileId = data.payload.project_facile_id;
    if (!projectFacileId) return;

    let project: { id: string } | undefined;
    for (let attempt = 0; attempt < 5; attempt++) {
      const [found] = await db
        .select({ id: projectTable.id })
        .from(projectTable)
        .where(
          and(
            eq(projectTable.facileId, projectFacileId),
            eq(projectTable.workspaceId, workspaceId),
          ),
        )
        .limit(1);

      if (found) {
        project = found;
        break;
      }
      await new Promise((r) => setTimeout(r, 1000));
    }

    if (!project) return;

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

    await db.insert(taskTable).values({
      projectId: project.id,
      title: data.payload.name,
      status,
      columnId: column?.id ?? null,
      facileId,
      number: nextNumber,
    });
  } else if (action === "updated") {
    const updateData: Record<string, unknown> = {};
    if (data.payload.name) updateData.title = data.payload.name;
    if (data.payload.status) {
      updateData.status = data.payload.status;

      const [task] = await db
        .select({ projectId: taskTable.projectId })
        .from(taskTable)
        .where(eq(taskTable.facileId, facileId))
        .limit(1);

      if (task) {
        const column = await db.query.columnTable.findFirst({
          where: and(
            eq(columnTable.projectId, task.projectId),
            eq(columnTable.slug, data.payload.status),
          ),
        });
        if (column) updateData.columnId = column.id;
      }
    }

    if (Object.keys(updateData).length > 0) {
      await db
        .update(taskTable)
        .set(updateData)
        .where(eq(taskTable.facileId, facileId));
    }
  } else if (action === "deleted") {
    await db.delete(taskTable).where(eq(taskTable.facileId, facileId));
  }
}

export function emitProjectEvent(
  workspaceId: string,
  action: "created" | "updated" | "deleted",
  project: {
    facileId: string | null;
    name: string;
    description?: string | null;
    icon?: string | null;
  },
): void {
  const client = activeClients.get(workspaceId);
  if (!client?.isConnected) return;

  const facileId = project.facileId;
  if (!facileId) return;

  client.emit(`project.${action}`, {
    app: "opus",
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

export function emitTaskEvent(
  workspaceId: string,
  action: "created" | "updated" | "deleted",
  task: {
    facileId: string | null;
    title: string;
    status: string;
    projectId: string;
  },
  projectFacileId: string | null,
): void {
  const client = activeClients.get(workspaceId);
  if (!client?.isConnected) return;

  const facileId = task.facileId;
  if (!facileId) return;

  client.emit(`task.${action}`, {
    app: "opus",
    object: "task",
    action,
    facile_id: facileId,
    payload: {
      facile_id: facileId,
      project_facile_id: projectFacileId || "",
      name: task.title,
      status: task.status,
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
