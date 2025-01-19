import Elysia, { t } from "elysia";
import { taskTable } from "../database/schema";

import { eq } from "drizzle-orm";
import db from "../database";
import getTasks from "./controllers/get-tasks";

const connections = new Map();

const task = new Elysia({ prefix: "/task" }).ws("/:projectId", {
  async open(ws) {
    const projectId = ws.data.params.projectId;

    // TODO: Improve this
    if (projectId === "undefined") {
      return null;
    }

    if (!connections.has(projectId)) {
      connections.set(projectId, new Set());
    }

    connections.get(projectId).add(ws);

    const boardState = await getTasks(projectId);
    ws.send(JSON.stringify(boardState));
  },
  async message(
    ws,
    message: {
      type: string;
      id: string;
      status: string;
    },
  ) {
    const projectId = ws.data.params.projectId;

    const data = message;

    if (data.type === "UPDATE_TASK") {
      // TODO: Make this part of updateTask
      await db
        .update(taskTable)
        .set({ status: data.status })
        .where(eq(taskTable.id, data.id));

      const clients = connections.get(projectId);
      const boardState = await getTasks(projectId);

      if (clients) {
        for (const client of clients) {
          client.send(JSON.stringify(boardState));
        }
      }
    }
  },
  close(ws) {
    const projectId = ws.data.params.projectId;

    const clients = connections.get(projectId);
    if (clients) {
      clients.delete(ws);

      if (clients.size === 0) {
        connections.delete(projectId);
      }
    }

    console.log(`Client disconnected from project ${projectId}`);
  },
});

export default task;
