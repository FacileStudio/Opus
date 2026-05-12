import { desc, eq } from "drizzle-orm";
import db, { schema } from "../../database";

export default async function getShareLinks(projectId: string) {
  return db
    .select()
    .from(schema.shareLinkTable)
    .where(eq(schema.shareLinkTable.projectId, projectId))
    .orderBy(desc(schema.shareLinkTable.createdAt));
}
