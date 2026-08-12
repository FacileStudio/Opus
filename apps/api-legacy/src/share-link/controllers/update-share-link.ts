import { eq } from "drizzle-orm";
import { HTTPException } from "hono/http-exception";
import db, { schema } from "../../database";

export default async function updateShareLink(
  id: string,
  data: {
    label?: string | null;
    expiresAt?: Date | null;
    isActive?: boolean;
  },
) {
  const [updated] = await db
    .update(schema.shareLinkTable)
    .set(data)
    .where(eq(schema.shareLinkTable.id, id))
    .returning();

  if (!updated) {
    throw new HTTPException(404, { message: "Share link not found" });
  }

  return updated;
}
