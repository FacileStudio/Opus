import { and, asc, eq } from "drizzle-orm";
import db from "../src/database";
import {
  columnTable,
  labelTable,
  projectTable,
  taskTable,
  userTable,
  workspaceTable,
  workspaceUserTable,
} from "../src/database/schema";

const DEFAULT_EMAIL = "aacevski@gmail.com";

type SeedTask = {
  title: string;
  description: string;
  status: "to-do" | "in-progress" | "in-review" | "done" | "planned" | "archived";
  priority: "low" | "medium" | "high";
  labels: string[];
  dueInDays?: number;
};

type SeedProject = {
  slug: string;
  name: string;
  icon: string;
  description: string;
  tasks: SeedTask[];
};

const DEFAULT_COLUMNS = [
  { isFinal: false, name: "To Do", position: 0, slug: "to-do" },
  { isFinal: false, name: "In Progress", position: 1, slug: "in-progress" },
  { isFinal: false, name: "In Review", position: 2, slug: "in-review" },
  { isFinal: true, name: "Done", position: 3, slug: "done" },
] as const;

const LABELS = [
  { color: "#EF4444", name: "Urgent" },
  { color: "#F59E0B", name: "Design" },
  { color: "#10B981", name: "Backend" },
  { color: "#3B82F6", name: "Frontend" },
  { color: "#8B5CF6", name: "Docs" },
  { color: "#06B6D4", name: "Infra" },
  { color: "#84CC16", name: "Growth" },
  { color: "#64748B", name: "Planned" },
] as const;

const PROJECTS: SeedProject[] = [
  {
    description: "Core product planning, execution, and release work.",
    icon: "Layout",
    name: "Platform Core",
    slug: "platform-core",
    tasks: [
      {
        description: "Finalize owner matrix, dependencies, and milestones for the quarter.",
        dueInDays: 5,
        labels: ["Urgent", "Growth"],
        priority: "high",
        status: "to-do",
        title: "Plan Q2 roadmap and release milestones",
      },
      {
        description: "Improve render throughput and drag responsiveness on large boards.",
        dueInDays: 8,
        labels: ["Frontend", "Urgent"],
        priority: "high",
        status: "in-progress",
        title: "Improve board performance with 1k+ tasks",
      },
      {
        description: "Clean up stale websocket subscriptions after workspace switches.",
        labels: ["Backend"],
        priority: "medium",
        status: "in-progress",
        title: "Fix real-time sync disconnect edge case",
      },
      {
        description: "Validate migration docs and rollback notes before release sign-off.",
        dueInDays: 3,
        labels: ["Docs"],
        priority: "medium",
        status: "in-review",
        title: "Review migration guide for v2",
      },
      {
        description: "Shipped completion tracking and default views for new users.",
        labels: ["Growth", "Frontend"],
        priority: "low",
        status: "done",
        title: "Ship onboarding completion funnel",
      },
      {
        description: "Investigate auto-assignment rule support for teams.",
        labels: ["Backend"],
        priority: "high",
        status: "planned",
        title: "Backlog: evaluate workflow auto-assignment rules",
      },
      {
        description: "Legacy checklist retained for reference only.",
        labels: ["Docs"],
        priority: "low",
        status: "archived",
        title: "Archive: v1 post-migration checklist",
      },
    ],
  },
  {
    description: "Website and conversion optimization work for kaneo.app.",
    icon: "Globe",
    name: "Marketing Site",
    slug: "marketing-site",
    tasks: [
      {
        description: "Align hero narrative, CTA hierarchy, and social proof placement.",
        dueInDays: 4,
        labels: ["Design", "Growth"],
        priority: "high",
        status: "to-do",
        title: "Refresh homepage information architecture",
      },
      {
        description: "Finalize responsive spacing and typography for launch pages.",
        labels: ["Design", "Frontend"],
        priority: "medium",
        status: "in-progress",
        title: "Polish responsive layout system",
      },
      {
        description: "Add sponsored-by section and link tracking attribution.",
        labels: ["Growth", "Frontend"],
        priority: "medium",
        status: "in-review",
        title: "Implement sponsorship placement in header",
      },
      {
        description: "Published copy and metadata updates for new release messaging.",
        labels: ["Growth", "Docs"],
        priority: "low",
        status: "done",
        title: "Ship release notes landing refresh",
      },
      {
        description: "Evaluate interactive product tour embed for hero section.",
        labels: ["Frontend"],
        priority: "medium",
        status: "planned",
        title: "Backlog: interactive product preview",
      },
    ],
  },
  {
    description: "Documentation IA, tutorials, and API reference quality.",
    icon: "Book",
    name: "Docs Experience",
    slug: "docs-experience",
    tasks: [
      {
        description: "Audit quick start path for first-time self-hosted teams.",
        dueInDays: 6,
        labels: ["Docs", "Urgent"],
        priority: "high",
        status: "to-do",
        title: "Restructure installation flow for time-to-first-project",
      },
      {
        description: "Consolidate workflow guides and remove duplicated examples.",
        labels: ["Docs"],
        priority: "medium",
        status: "in-progress",
        title: "Unify functional guides under one narrative",
      },
      {
        description: "Verify API examples against current schema and responses.",
        labels: ["Docs", "Backend"],
        priority: "medium",
        status: "in-review",
        title: "Review API reference code snippets",
      },
      {
        description: "Published migration FAQ and troubleshooting links.",
        labels: ["Docs"],
        priority: "low",
        status: "done",
        title: "Ship migration FAQ update",
      },
      {
        description: "Plan walkthrough videos for common onboarding flows.",
        labels: ["Docs", "Growth"],
        priority: "low",
        status: "planned",
        title: "Backlog: add video onboarding modules",
      },
    ],
  },
  {
    description: "GitHub app integration reliability and synchronization.",
    icon: "Github",
    name: "GitHub Sync",
    slug: "github-sync",
    tasks: [
      {
        description: "Stabilize webhook retries and event deduplication logic.",
        dueInDays: 7,
        labels: ["Backend", "Infra", "Urgent"],
        priority: "high",
        status: "to-do",
        title: "Harden webhook delivery and replay safety",
      },
      {
        description: "Improve status transition mapping when PR merges rapidly.",
        labels: ["Backend"],
        priority: "medium",
        status: "in-progress",
        title: "Refine issue-to-task state transitions",
      },
      {
        description: "Validate installation/setup docs with fresh org test account.",
        labels: ["Docs", "Backend"],
        priority: "medium",
        status: "in-review",
        title: "Review GitHub integration setup path",
      },
      {
        description: "Released fix for missing external link metadata records.",
        labels: ["Backend"],
        priority: "low",
        status: "done",
        title: "Ship external link metadata patch",
      },
      {
        description: "Investigate bi-directional label sync with conflict handling.",
        labels: ["Backend", "Planned"],
        priority: "medium",
        status: "planned",
        title: "Backlog: bi-directional label sync",
      },
      {
        description: "Deprecated old OAuth fallback path after app rollout.",
        labels: ["Backend"],
        priority: "low",
        status: "archived",
        title: "Archive: legacy OAuth fallback support",
      },
    ],
  },
];

function slugify(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

async function ensureWorkspaceForUser(userId: string, email: string) {
  const localPart = slugify(email.split("@")[0] ?? "user");
  const workspaceSlug = `${localPart}-workspace`;
  const workspaceName = "Kaneo Product Studio";

  let workspace = await db.query.workspaceTable.findFirst({
    where: eq(workspaceTable.slug, workspaceSlug),
  });

  if (!workspace) {
    [workspace] = await db
      .insert(workspaceTable)
      .values({
        createdAt: new Date(),
        description: "Product, docs, site, and integration planning workspace.",
        name: workspaceName,
        slug: workspaceSlug,
      })
      .returning();
    console.log(`Created workspace: ${workspace.name} (${workspace.slug})`);
  } else {
    await db
      .update(workspaceTable)
      .set({
        description: "Product, docs, site, and integration planning workspace.",
        name: workspaceName,
      })
      .where(eq(workspaceTable.id, workspace.id));
    console.log(`Using workspace: ${workspace.name} (${workspace.slug})`);
  }

  const membership = await db.query.workspaceUserTable.findFirst({
    where: and(
      eq(workspaceUserTable.userId, userId),
      eq(workspaceUserTable.workspaceId, workspace.id),
    ),
  });

  if (!membership) {
    await db.insert(workspaceUserTable).values({
      joinedAt: new Date(),
      role: "owner",
      userId,
      workspaceId: workspace.id,
    });
    console.log("Created workspace membership");
  }

  return workspace;
}

async function ensureProject(workspaceId: string, seedProject: SeedProject) {
  let project = await db.query.projectTable.findFirst({
    where: and(
      eq(projectTable.workspaceId, workspaceId),
      eq(projectTable.slug, seedProject.slug),
    ),
  });

  if (!project) {
    [project] = await db
      .insert(projectTable)
      .values({
        description: seedProject.description,
        icon: seedProject.icon,
        name: seedProject.name,
        slug: seedProject.slug,
        workspaceId,
      })
      .returning();
    console.log(`Created project: ${project.name}`);
  } else {
    await db
      .update(projectTable)
      .set({
        description: seedProject.description,
        icon: seedProject.icon,
        name: seedProject.name,
      })
      .where(eq(projectTable.id, project.id));
    console.log(`Using project: ${project.name}`);
  }

  const existingColumns = await db
    .select()
    .from(columnTable)
    .where(eq(columnTable.projectId, project.id))
    .orderBy(asc(columnTable.position));

  const existingBySlug = new Set(existingColumns.map((col) => col.slug));

  for (const column of DEFAULT_COLUMNS) {
    if (!existingBySlug.has(column.slug)) {
      await db.insert(columnTable).values({
        isFinal: column.isFinal,
        name: column.name,
        position: column.position,
        projectId: project.id,
        slug: column.slug,
      });
      console.log(`Created column ${column.name} for ${seedProject.name}`);
    }
  }

  return project;
}

async function ensureTasks(projectId: string, userId: string, seedTasks: SeedTask[]) {
  const projectColumns = await db
    .select()
    .from(columnTable)
    .where(eq(columnTable.projectId, projectId));
  const columnIdBySlug = new Map(projectColumns.map((col) => [col.slug, col.id]));

  const existingTasks = await db
    .select()
    .from(taskTable)
    .where(eq(taskTable.projectId, projectId));

  let nextNumber = existingTasks.reduce((max, task) => Math.max(max, task.number ?? 0), 0);
  let nextPosition = existingTasks.reduce((max, task) => Math.max(max, task.position ?? 0), 0);

  for (const seedTask of seedTasks) {
    const alreadyExists = existingTasks.some((task) => task.title === seedTask.title);
    if (alreadyExists) continue;

    nextNumber += 1;
    nextPosition += 1;

    const dueDate = seedTask.dueInDays
      ? new Date(Date.now() + seedTask.dueInDays * 24 * 60 * 60 * 1000)
      : null;

    await db.insert(taskTable).values({
      columnId:
        seedTask.status === "planned" || seedTask.status === "archived"
          ? null
          : (columnIdBySlug.get(seedTask.status) ?? null),
      description: seedTask.description,
      dueDate,
      number: nextNumber,
      position: nextPosition,
      priority: seedTask.priority,
      projectId,
      status: seedTask.status,
      title: seedTask.title,
      userId,
    });
    console.log(`Created task: ${seedTask.title}`);
  }

  return db
    .select()
    .from(taskTable)
    .where(eq(taskTable.projectId, projectId));
}

async function ensureWorkspaceLabels(workspaceId: string) {
  for (const workspaceLabel of LABELS) {
    const exists = await db.query.labelTable.findFirst({
      where: and(
        eq(labelTable.workspaceId, workspaceId),
        eq(labelTable.name, workspaceLabel.name),
        eq(labelTable.color, workspaceLabel.color),
      ),
    });

    if (!exists) {
      await db.insert(labelTable).values({
        color: workspaceLabel.color,
        name: workspaceLabel.name,
        workspaceId,
      });
      console.log(`Created workspace label: ${workspaceLabel.name}`);
    }
  }

  return db
    .select()
    .from(labelTable)
    .where(eq(labelTable.workspaceId, workspaceId));
}

async function ensureTaskLabels(
  workspaceId: string,
  projectTasks: Array<typeof taskTable.$inferSelect>,
  seedTasks: SeedTask[],
  workspaceLabels: Array<typeof labelTable.$inferSelect>,
) {
  for (const seedTask of seedTasks) {
    const task = projectTasks.find((t) => t.title === seedTask.title);
    if (!task) continue;

    for (const labelName of seedTask.labels) {
      const labelTemplate = workspaceLabels.find((label) => label.name === labelName);
      if (!labelTemplate) continue;

      const exists = await db.query.labelTable.findFirst({
        where: and(
          eq(labelTable.workspaceId, workspaceId),
          eq(labelTable.taskId, task.id),
          eq(labelTable.name, labelTemplate.name),
          eq(labelTable.color, labelTemplate.color),
        ),
      });

      if (exists) continue;

      await db.insert(labelTable).values({
        color: labelTemplate.color,
        name: labelTemplate.name,
        taskId: task.id,
        workspaceId,
      });
      console.log(`Attached label ${labelTemplate.name} to task ${task.title}`);
    }
  }
}

async function run() {
  const targetEmail = process.argv[2] ?? DEFAULT_EMAIL;
  console.log(`Seeding demo data for ${targetEmail}`);

  const user = await db.query.userTable.findFirst({
    where: eq(userTable.email, targetEmail),
  });

  if (!user) {
    throw new Error(`User not found: ${targetEmail}. Create/sign-in this user first.`);
  }

  const workspace = await ensureWorkspaceForUser(user.id, targetEmail);
  const workspaceLabels = await ensureWorkspaceLabels(workspace.id);

  for (const seedProject of PROJECTS) {
    const project = await ensureProject(workspace.id, seedProject);
    const projectTasks = await ensureTasks(project.id, user.id, seedProject.tasks);
    await ensureTaskLabels(workspace.id, projectTasks, seedProject.tasks, workspaceLabels);
  }

  console.log("Seed completed successfully.");
}

run()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error("Seed failed:", error);
    process.exit(1);
  });
