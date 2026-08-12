import { eq } from "drizzle-orm";
import { HTTPException } from "hono/http-exception";
import db, { schema } from "../../database";

export default async function deleteShareLink(id: string) {
  const [deleted] = await db
    .delete(schema.shareLinkTable)
    .where(eq(schema.shareLinkTable.id, id))
    .returning();

  if (!deleted) {
    throw new HTTPException(404, { message: "Share link not found" });
  }

  return deleted;
}
