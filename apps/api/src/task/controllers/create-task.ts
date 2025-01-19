import db from "../../database";
import { taskTable } from "../../database/schema";

async function createTask(body: {
  projectId: string;
  assigneeId: string;
  title: string;
  status: string;
  dueDate: Date;
  description: string;
}) {
  return db.insert(taskTable).values(body);
}

export default createTask;
