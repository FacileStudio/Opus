import { eq, sql } from "drizzle-orm";
import { HTTPException } from "hono/http-exception";
import db, { schema } from "../../database";
import getTasks from "../../task/controllers/get-tasks";

export default async function getSharedProject(token: string) {
  const [shareLink] = await db
    .select()
    .from(schema.shareLinkTable)
    .where(eq(schema.shareLinkTable.token, token))
    .limit(1);

  if (!shareLink) {
    throw new HTTPException(404, { message: "Not found" });
  }

  if (!shareLink.isActive) {
    throw new HTTPException(404, { message: "Not found" });
  }

  if (shareLink.expiresAt && shareLink.expiresAt < new Date()) {
    throw new HTTPException(404, { message: "Not found" });
  }

  await db
    .update(schema.shareLinkTable)
    .set({
      lastAccessedAt: new Date(),
      accessCount: sql`${schema.shareLinkTable.accessCount} + 1`,
    })
    .where(eq(schema.shareLinkTable.id, shareLink.id));

  const result = await getTasks(shareLink.projectId);

  if (!result.data) {
    throw new HTTPException(404, { message: "Project not found" });
  }

  const { workspaceId: _, ...safeData } = result.data;
  return safeData;
}
