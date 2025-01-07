import { createInsertSchema } from "drizzle-typebox";
import { t } from "elysia";
import { workspaceTable } from "../../database/schema";

export const createWorkspaceSchema = createInsertSchema(workspaceTable, {
  name: t.String(),
});
