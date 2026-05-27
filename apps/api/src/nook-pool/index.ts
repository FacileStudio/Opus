import { eq } from "drizzle-orm";
import { Hono } from "hono";
import { HTTPException } from "hono/http-exception";
import { describeRoute, resolver, validator } from "hono-openapi";
import * as v from "valibot";
import db, { schema } from "../database";
import { workspaceAccess } from "../utils/workspace-access-middleware";
import { connectPoolForWorkspace, disconnectPoolForWorkspace } from "./sync";

const nookPoolConfigSchema = v.object({
  id: v.string(),
  workspaceId: v.string(),
  instanceUrl: v.string(),
  enabled: v.boolean(),
  createdAt: v.string(),
  updatedAt: v.string(),
});

const nookPool = new Hono<{
  Variables: {
    userId: string;
    workspaceId: string;
  };
}>()
  .get(
    "/:workspaceId",
    describeRoute({
      operationId: "getNookPoolConfig",
      tags: ["Nook Pool"],
      description: "Get Nook Pool configuration for a workspace",
      responses: {
        200: {
          description: "Nook Pool configuration",
          content: {
            "application/json": {
              schema: resolver(v.nullable(nookPoolConfigSchema)),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const [config] = await db
        .select({
          id: schema.nookPoolConfigTable.id,
          workspaceId: schema.nookPoolConfigTable.workspaceId,
          instanceUrl: schema.nookPoolConfigTable.instanceUrl,
          enabled: schema.nookPoolConfigTable.enabled,
          createdAt: schema.nookPoolConfigTable.createdAt,
          updatedAt: schema.nookPoolConfigTable.updatedAt,
        })
        .from(schema.nookPoolConfigTable)
        .where(eq(schema.nookPoolConfigTable.workspaceId, workspaceId))
        .limit(1);

      return c.json(config ?? null);
    },
  )
  .post(
    "/:workspaceId",
    describeRoute({
      operationId: "saveNookPoolConfig",
      tags: ["Nook Pool"],
      description: "Save Nook Pool configuration for a workspace",
      responses: {
        200: {
          description: "Nook Pool configuration saved",
          content: {
            "application/json": {
              schema: resolver(nookPoolConfigSchema),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    validator(
      "json",
      v.object({
        instanceUrl: v.pipe(v.string(), v.url()),
        secret: v.string(),
        enabled: v.optional(v.boolean()),
      }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const { instanceUrl, secret, enabled } = c.req.valid("json");

      const [existing] = await db
        .select({ id: schema.nookPoolConfigTable.id })
        .from(schema.nookPoolConfigTable)
        .where(eq(schema.nookPoolConfigTable.workspaceId, workspaceId))
        .limit(1);

      const config = existing
        ? (
            await db
              .update(schema.nookPoolConfigTable)
              .set({ instanceUrl, secret, enabled: enabled ?? true })
              .where(eq(schema.nookPoolConfigTable.id, existing.id))
              .returning()
          )[0]
        : (
            await db
              .insert(schema.nookPoolConfigTable)
              .values({
                workspaceId,
                instanceUrl,
                secret,
                enabled: enabled ?? true,
              })
              .returning()
          )[0];

      if (!config) {
        throw new HTTPException(500, {
          message: "Failed to save Nook Pool config",
        });
      }

      if (config.enabled) {
        try {
          await connectPoolForWorkspace(workspaceId, instanceUrl, secret);
        } catch (error) {
          console.error(
            `Failed to connect Pool for workspace ${workspaceId}:`,
            error,
          );
        }
      } else {
        disconnectPoolForWorkspace(workspaceId);
      }

      return c.json({
        id: config.id,
        workspaceId: config.workspaceId,
        instanceUrl: config.instanceUrl,
        enabled: config.enabled,
        createdAt: config.createdAt,
        updatedAt: config.updatedAt,
      });
    },
  )
  .delete(
    "/:workspaceId",
    describeRoute({
      operationId: "deleteNookPoolConfig",
      tags: ["Nook Pool"],
      description: "Delete Nook Pool configuration and disconnect",
      responses: {
        200: {
          description: "Nook Pool configuration deleted",
          content: {
            "application/json": {
              schema: resolver(v.object({ success: v.boolean() })),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");

      disconnectPoolForWorkspace(workspaceId);

      const deleted = await db
        .delete(schema.nookPoolConfigTable)
        .where(eq(schema.nookPoolConfigTable.workspaceId, workspaceId))
        .returning();

      if (!deleted.length) {
        throw new HTTPException(404, {
          message: "Nook Pool config not found",
        });
      }

      return c.json({ success: true });
    },
  )
  .post(
    "/:workspaceId/test",
    describeRoute({
      operationId: "testNookPoolConnection",
      tags: ["Nook Pool"],
      description: "Test connection to a Nook Pool instance",
      responses: {
        200: {
          description: "Connection test result",
          content: {
            "application/json": {
              schema: resolver(
                v.object({
                  success: v.boolean(),
                  error: v.optional(v.string()),
                }),
              ),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    validator(
      "json",
      v.object({
        instanceUrl: v.pipe(v.string(), v.url()),
        secret: v.string(),
      }),
    ),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const { instanceUrl, secret } = c.req.valid("json");

      try {
        const url = `${instanceUrl.replace(/\/+$/, "")}/api/pool/register`;
        const res = await fetch(url, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            app: "opus",
            secret,
            events: { emit: [], listen: [] },
          }),
          signal: AbortSignal.timeout(10000),
        });

        if (!res.ok) {
          const text = await res.text();
          return c.json({
            success: false,
            error: `Registration failed: ${res.status} ${text}`,
          });
        }

        return c.json({ success: true });
      } catch (err) {
        return c.json({
          success: false,
          error: err instanceof Error ? err.message : "Connection failed",
        });
      }
    },
  );

export default nookPool;
