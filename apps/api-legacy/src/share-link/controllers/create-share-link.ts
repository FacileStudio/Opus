import db, { schema } from "../../database";

export default async function createShareLink(
  projectId: string,
  createdBy: string,
  label?: string,
  expiresAt?: Date,
) {
  const [shareLink] = await db
    .insert(schema.shareLinkTable)
    .values({
      projectId,
      createdBy,
      label: label || null,
      expiresAt: expiresAt || null,
    })
    .returning();

  return shareLink;
}
