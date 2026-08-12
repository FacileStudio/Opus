import { Hono } from "hono";
import { HTTPException } from "hono/http-exception";
import { describeRoute, resolver, validator } from "hono-openapi";
import * as v from "valibot";
import { shareLinkSchema } from "../schemas";
import { workspaceAccess } from "../utils/workspace-access-middleware";
import createShareLinkCtrl from "./controllers/create-share-link";
import deleteShareLinkCtrl from "./controllers/delete-share-link";
import getShareLinksCtrl from "./controllers/get-share-links";
import updateShareLinkCtrl from "./controllers/update-share-link";

const shareLink = new Hono<{
  Variables: {
    userId: string;
    workspaceId: string;
  };
}>()
  .get(
    "/",
    describeRoute({
      operationId: "listShareLinks",
      tags: ["Share Links"],
      description: "List all share links for a project",
      responses: {
        200: {
          description: "List of share links",
          content: {
            "application/json": {
              schema: resolver(v.array(shareLinkSchema)),
            },
          },
        },
      },
    }),
    validator(
      "query",
      v.object({
        projectId: v.string(),
      }),
    ),
    workspaceAccess.fromProject("projectId"),
    async (c) => {
      const { projectId } = c.req.valid("query");
      const links = await getShareLinksCtrl(projectId);
      return c.json(links);
    },
  )
  .post(
    "/",
    describeRoute({
      operationId: "createShareLink",
      tags: ["Share Links"],
      description: "Create a new share link for a project",
      responses: {
        200: {
          description: "Share link created",
          content: {
            "application/json": { schema: resolver(shareLinkSchema) },
          },
        },
      },
    }),
    validator(
      "json",
      v.object({
        projectId: v.string(),
        label: v.optional(v.string()),
        expiresAt: v.optional(v.string()),
      }),
    ),
    workspaceAccess.fromProject("projectId"),
    async (c) => {
      const userId = c.get("userId");
      const { projectId, label, expiresAt } = c.req.valid("json");
      let parsedExpiry: Date | undefined;
      if (expiresAt) {
        parsedExpiry = new Date(expiresAt);
        if (Number.isNaN(parsedExpiry.getTime())) {
          throw new HTTPException(400, { message: "Invalid expiry date" });
        }
        if (parsedExpiry <= new Date()) {
          throw new HTTPException(400, {
            message: "Expiry date must be in the future",
          });
        }
      }
      const link = await createShareLinkCtrl(
        projectId,
        userId,
        label,
        parsedExpiry,
      );
      return c.json(link);
    },
  )
  .put(
    "/:id",
    describeRoute({
      operationId: "updateShareLink",
      tags: ["Share Links"],
      description: "Update a share link",
      responses: {
        200: {
          description: "Share link updated",
          content: {
            "application/json": { schema: resolver(shareLinkSchema) },
          },
        },
      },
    }),
    validator("param", v.object({ id: v.string() })),
    validator(
      "json",
      v.object({
        label: v.optional(v.nullable(v.string())),
        expiresAt: v.optional(v.nullable(v.string())),
        isActive: v.optional(v.boolean()),
      }),
    ),
    workspaceAccess.fromShareLink(),
    async (c) => {
      const { id } = c.req.valid("param");
      const body = c.req.valid("json");
      const data: {
        label?: string | null;
        expiresAt?: Date | null;
        isActive?: boolean;
      } = {};
      if (body.label !== undefined) data.label = body.label;
      if (body.expiresAt !== undefined) {
        if (body.expiresAt) {
          const parsed = new Date(body.expiresAt);
          if (Number.isNaN(parsed.getTime())) {
            throw new HTTPException(400, { message: "Invalid expiry date" });
          }
          data.expiresAt = parsed;
        } else {
          data.expiresAt = null;
        }
      }
      if (body.isActive !== undefined) data.isActive = body.isActive;
      const link = await updateShareLinkCtrl(id, data);
      return c.json(link);
    },
  )
  .delete(
    "/:id",
    describeRoute({
      operationId: "deleteShareLink",
      tags: ["Share Links"],
      description: "Delete a share link",
      responses: {
        200: {
          description: "Share link deleted",
          content: {
            "application/json": { schema: resolver(shareLinkSchema) },
          },
        },
      },
    }),
    validator("param", v.object({ id: v.string() })),
    workspaceAccess.fromShareLink(),
    async (c) => {
      const { id } = c.req.valid("param");
      const deleted = await deleteShareLinkCtrl(id);
      return c.json(deleted);
    },
  );

export default shareLink;
