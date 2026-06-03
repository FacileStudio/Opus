import { Hono } from "hono";
import { bodyLimit } from "hono/body-limit";
import { HTTPException } from "hono/http-exception";
import { describeRoute, resolver, validator } from "hono-openapi";
import * as v from "valibot";
import {
  deleteWorkspaceLogo,
  extensionFromMime,
  saveWorkspaceLogo,
  validateLogoUpload,
} from "../storage/local";
import { workspaceAccess } from "../utils/workspace-access-middleware";
import getUserWorkspacesCtrl from "./controllers/get-user-workspaces";
import getWorkspaceMembersCtrl from "./controllers/get-workspace-members";

const workspace = new Hono<{
  Variables: {
    userId: string;
    workspaceId: string;
  };
}>()
  .get(
    "/",
    describeRoute({
      operationId: "getUserWorkspaces",
      tags: ["Workspaces"],
      description: "Get all workspaces the authenticated user belongs to",
      responses: {
        200: {
          description: "List of workspaces",
          content: {
            "application/json": {
              schema: resolver(
                v.array(
                  v.object({
                    id: v.string(),
                    name: v.string(),
                    slug: v.string(),
                    logo: v.nullable(v.string()),
                    description: v.nullable(v.string()),
                    createdAt: v.string(),
                  }),
                ),
              ),
            },
          },
        },
      },
    }),
    async (c) => {
      const userId = c.get("userId");
      const workspaces = await getUserWorkspacesCtrl(userId);
      return c.json(workspaces);
    },
  )
  .get(
    "/:workspaceId/members",
    describeRoute({
      operationId: "getWorkspaceMembers",
      tags: ["Workspaces"],
      description: "Get all members of a workspace",
      responses: {
        200: {
          description: "List of workspace members",
          content: {
            "application/json": {
              schema: resolver(
                v.array(
                  v.object({
                    id: v.string(),
                    name: v.string(),
                    email: v.string(),
                    image: v.nullable(v.string()),
                    role: v.string(),
                  }),
                ),
              ),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const members = await getWorkspaceMembersCtrl(workspaceId);
      return c.json(members);
    },
  );

const workspaceLogoUpload = new Hono<{
  Variables: {
    userId: string;
    workspaceId: string;
  };
}>()
  .post(
    "/:workspaceId/logo",
    describeRoute({
      operationId: "uploadWorkspaceLogo",
      tags: ["Workspaces"],
      description: "Upload a logo image for a workspace",
      responses: {
        200: {
          description: "Logo uploaded successfully",
          content: {
            "application/json": {
              schema: resolver(
                v.object({ url: v.string(), message: v.string() }),
              ),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    bodyLimit({
      maxSize: 2 * 1024 * 1024,
      onError: () => {
        throw new HTTPException(413, {
          message: "File too large. Maximum 2MB",
        });
      },
    }),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      const body = await c.req.parseBody();
      const file = body.logo;

      if (!(file instanceof File)) {
        throw new HTTPException(400, {
          message:
            "No logo file provided. Send as multipart with field name 'logo'",
        });
      }

      validateLogoUpload(file.type, file.size);

      const buffer = Buffer.from(await file.arrayBuffer());
      const ext = extensionFromMime(file.type);

      await saveWorkspaceLogo(workspaceId, buffer, ext);

      const logoUrl = `/api/workspace/${workspaceId}/logo?v=${Date.now()}`;

      return c.json({ url: logoUrl, message: "Logo uploaded successfully" });
    },
  )
  .delete(
    "/:workspaceId/logo",
    describeRoute({
      operationId: "deleteWorkspaceLogo",
      tags: ["Workspaces"],
      description: "Delete a workspace logo",
      responses: {
        200: {
          description: "Logo deleted successfully",
          content: {
            "application/json": {
              schema: resolver(v.object({ message: v.string() })),
            },
          },
        },
      },
    }),
    validator("param", v.object({ workspaceId: v.string() })),
    workspaceAccess.fromParam("workspaceId"),
    async (c) => {
      const workspaceId = c.get("workspaceId");
      await deleteWorkspaceLogo(workspaceId);
      return c.json({ message: "Logo deleted" });
    },
  );

export { workspaceLogoUpload };
export default workspace;
