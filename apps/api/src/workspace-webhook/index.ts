import { and, eq } from "drizzle-orm";
import { Hono } from "hono";
import { HTTPException } from "hono/http-exception";
import { describeRoute, resolver, validator } from "hono-openapi";
import * as v from "valibot";
import db, { schema } from "../database";
import { workspaceAccess } from "../utils/workspace-access-middleware";

const webhookSchema = v.object({
  id: v.string(),
  workspaceId: v.string(),
  name: v.string(),
  url: v.string(),
  events: v.string(),
  enabled: v.boolean(),
  createdAt: v.string(),
  updatedAt: v.string(),
});

const workspaceWebhook = new Hono<{
  Variables: {
    userId: string;
    workspaceId: string;
  };
}>();

workspaceWebhook
  .get(
    "/:workspaceId",
    describeRoute({
      operationId: "listWorkspaceWebhooks",
      tags: ["Workspace Webhooks"],
      description: "List all webhooks for a workspace",
      responses: {
        200: {
          description: "List of workspace webhooks",
          content: {
            "application/json": {
              schema: resolver(v.array(webhookSchema)),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const webhooks = await db
        .select({
          id: schema.workspaceWebhookTable.id,
          workspaceId: schema.workspaceWebhookTable.workspaceId,
          name: schema.workspaceWebhookTable.name,
          url: schema.workspaceWebhookTable.url,
          events: schema.workspaceWebhookTable.events,
          enabled: schema.workspaceWebhookTable.enabled,
          createdAt: schema.workspaceWebhookTable.createdAt,
          updatedAt: schema.workspaceWebhookTable.updatedAt,
        })
        .from(schema.workspaceWebhookTable)
        .where(eq(schema.workspaceWebhookTable.workspaceId, workspaceId));
      return c.json(webhooks);
    },
  )
  .post(
    "/:workspaceId",
    describeRoute({
      operationId: "createWorkspaceWebhook",
      tags: ["Workspace Webhooks"],
      description: "Create a new webhook for a workspace",
      responses: {
        200: {
          description: "Created webhook",
          content: {
            "application/json": {
              schema: resolver(webhookSchema),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    validator(
      "json",
      v.object({
        name: v.string(),
        url: v.pipe(v.string(), v.url()),
        secret: v.optional(v.nullable(v.string())),
        events: v.optional(v.string()),
        enabled: v.optional(v.boolean()),
      }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const body = c.req.valid("json");
      const [webhook] = await db
        .insert(schema.workspaceWebhookTable)
        .values({
          workspaceId,
          name: body.name,
          url: body.url,
          secret: body.secret ?? null,
          events: body.events ?? "*",
          enabled: body.enabled ?? true,
        })
        .returning();
      return c.json(webhook);
    },
  )
  .patch(
    "/:workspaceId/:webhookId",
    describeRoute({
      operationId: "updateWorkspaceWebhook",
      tags: ["Workspace Webhooks"],
      description: "Update an existing workspace webhook",
      responses: {
        200: {
          description: "Updated webhook",
          content: {
            "application/json": {
              schema: resolver(webhookSchema),
            },
          },
        },
      },
    }),
    validator(
      "param",
      v.object({ workspaceId: v.string(), webhookId: v.string() }),
    ),
    validator(
      "json",
      v.object({
        name: v.optional(v.string()),
        url: v.optional(v.pipe(v.string(), v.url())),
        secret: v.optional(v.nullable(v.string())),
        events: v.optional(v.string()),
        enabled: v.optional(v.boolean()),
      }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const { webhookId } = c.req.valid("param");
      const workspaceId = c.get("workspaceId");
      const body = c.req.valid("json");
      const [webhook] = await db
        .update(schema.workspaceWebhookTable)
        .set(body)
        .where(
          and(
            eq(schema.workspaceWebhookTable.id, webhookId),
            eq(schema.workspaceWebhookTable.workspaceId, workspaceId),
          ),
        )
        .returning();
      if (!webhook) {
        throw new HTTPException(404, { message: "Webhook not found" });
      }
      return c.json(webhook);
    },
  )
  .delete(
    "/:workspaceId/:webhookId",
    describeRoute({
      operationId: "deleteWorkspaceWebhook",
      tags: ["Workspace Webhooks"],
      description: "Delete a workspace webhook",
      responses: {
        200: {
          description: "Deleted",
          content: {
            "application/json": {
              schema: resolver(v.object({ success: v.boolean() })),
            },
          },
        },
      },
    }),
    validator(
      "param",
      v.object({ workspaceId: v.string(), webhookId: v.string() }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const { webhookId } = c.req.valid("param");
      const workspaceId = c.get("workspaceId");
      const deleted = await db
        .delete(schema.workspaceWebhookTable)
        .where(
          and(
            eq(schema.workspaceWebhookTable.id, webhookId),
            eq(schema.workspaceWebhookTable.workspaceId, workspaceId),
          ),
        )
        .returning();
      if (!deleted.length) {
        throw new HTTPException(404, { message: "Webhook not found" });
      }
      return c.json({ success: true });
    },
  )
  .post(
    "/:workspaceId/:webhookId/test",
    describeRoute({
      operationId: "testWorkspaceWebhook",
      tags: ["Workspace Webhooks"],
      description: "Send a test ping to a workspace webhook",
      responses: {
        200: {
          description: "Test result",
          content: {
            "application/json": {
              schema: resolver(
                v.object({
                  success: v.boolean(),
                  statusCode: v.optional(v.number()),
                  error: v.optional(v.string()),
                }),
              ),
            },
          },
        },
      },
    }),
    validator(
      "param",
      v.object({ workspaceId: v.string(), webhookId: v.string() }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const { webhookId } = c.req.valid("param");
      const workspaceId = c.get("workspaceId");
      const [webhook] = await db
        .select()
        .from(schema.workspaceWebhookTable)
        .where(
          and(
            eq(schema.workspaceWebhookTable.id, webhookId),
            eq(schema.workspaceWebhookTable.workspaceId, workspaceId),
          ),
        )
        .limit(1);
      if (!webhook) {
        throw new HTTPException(404, { message: "Webhook not found" });
      }
      try {
        const payload = JSON.stringify({
          event: "ping",
          timestamp: new Date().toISOString(),
          webhookId: webhook.id,
          workspaceId: webhook.workspaceId,
        });
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
          "X-Opus-Event": "ping",
          "X-Opus-Webhook-Id": webhook.id,
        };
        if (webhook.secret) {
          const encoder = new TextEncoder();
          const key = await crypto.subtle.importKey(
            "raw",
            encoder.encode(webhook.secret),
            { name: "HMAC", hash: "SHA-256" },
            false,
            ["sign"],
          );
          const signature = await crypto.subtle.sign(
            "HMAC",
            key,
            encoder.encode(payload),
          );
          const hexSignature = Array.from(new Uint8Array(signature))
            .map((b) => b.toString(16).padStart(2, "0"))
            .join("");
          headers["X-Opus-Signature"] = `sha256=${hexSignature}`;
        }
        const response = await fetch(webhook.url, {
          method: "POST",
          headers,
          body: payload,
          signal: AbortSignal.timeout(10000),
        });
        return c.json({
          success: response.ok,
          statusCode: response.status,
        });
      } catch (err) {
        return c.json({
          success: false,
          error: err instanceof Error ? err.message : "Unknown error",
        });
      }
    },
  );

export default workspaceWebhook;
