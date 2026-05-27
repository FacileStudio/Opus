import { PoolClient } from "@facile/pool";
import { createId } from "@paralleldrive/cuid2";
import { and, eq, isNull } from "drizzle-orm";
import db from "../database";
import {
  columnTable,
  nookPoolConfigTable,
  projectTable,
  taskTable,
} from "../database/schema";

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

function setupListeners(pool: PoolClient, workspaceId: string): void {
  pool.listen("project.created", async (payload: unknown) => {
    try {
      await handleIncomingProject("created", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle project.created", error);
    }
  });

  pool.listen("project.updated", async (payload: unknown) => {
    try {
      await handleIncomingProject("updated", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle project.updated", error);
    }
  });

  pool.listen("project.deleted", async (payload: unknown) => {
    try {
      await handleIncomingProject("deleted", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle project.deleted", error);
    }
  });

  pool.listen("task.created", async (payload: unknown) => {
    try {
      await handleIncomingTask("created", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle task.created", error);
    }
  });

  pool.listen("task.updated", async (payload: unknown) => {
    try {
      await handleIncomingTask("updated", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle task.updated", error);
    }
  });

  pool.listen("task.deleted", async (payload: unknown) => {
    try {
      await handleIncomingTask("deleted", workspaceId, payload);
    } catch (error) {
      console.error("Pool: failed to handle task.deleted", error);
    }
  });
}

type IncomingProjectPayload = {
  facile_id: string;
  payload: {
    facile_id: string;
    name: string;
    description?: string;
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

    await db.insert(projectTable).values({
      workspaceId,
      name: data.payload.name,
      description: data.payload.description || null,
      slug,
      facileId,
    });
  } else if (action === "updated") {
    const updateData: Record<string, unknown> = {};
    if (data.payload.name) updateData.name = data.payload.name;
    if (data.payload.description !== undefined)
      updateData.description = data.payload.description;

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

    if (!project) return;

    const status = data.payload.status || "to-do";

    const column = await db.query.columnTable.findFirst({
      where: and(
        eq(columnTable.projectId, project.id),
        eq(columnTable.slug, status),
      ),
    });

    await db.insert(taskTable).values({
      projectId: project.id,
      title: data.payload.name,
      status,
      columnId: column?.id ?? null,
      facileId,
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
    },
    timestamp: Date.now(),
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
    timestamp: Date.now(),
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
