import { addIcon } from "@iconify/react";

export const IDEAS_ICON = "solar:lightbulb-linear";

addIcon(IDEAS_ICON, {
  body: '<g fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="1.5"><path d="M14.5 19.5H9.5M14.5 19.5C14.5 18.7865 14.5 18.4297 14.5381 18.193C14.6609 17.4296 14.6824 17.3815 15.1692 16.7807C15.3201 16.5945 15.8805 16.0927 17.0012 15.0892C18.5349 13.7159 19.5 11.7206 19.5 9.5C19.5 5.35786 16.1421 2 12 2C7.85786 2 4.5 5.35786 4.5 9.5C4.5 11.7206 5.4651 13.7159 6.99876 15.0892C8.11945 16.0927 8.67987 16.5945 8.83082 16.7807C9.31762 17.3815 9.3391 17.4296 9.46192 18.193C9.5 18.4297 9.5 18.7865 9.5 19.5M14.5 19.5C14.5 20.4346 14.5 20.9019 14.299 21.25C14.1674 21.478 13.978 21.6674 13.75 21.799C13.4019 22 12.9346 22 12 22C11.0654 22 10.5981 22 10.25 21.799C10.022 21.6674 9.83261 21.478 9.70096 21.25C9.5 20.9019 9.5 20.4346 9.5 19.5"/><path d="M12 17V15"/><path d="M13.7324 14C13.3866 14.5978 12.7403 15 12 15C11.2597 15 10.6134 14.5978 10.2676 14"/></g>',
  width: 24,
  height: 24,
});

export type IdeaSource = { label: string; ref: string };

export type Idea = {
  id: string;
  kind: string;
  title: string;
  why: string;
  draft: string;
  sources: IdeaSource[] | null;
  status: string;
  created_at: string;
  expires_at: string;
};

export type IdeasPageData = {
  ai_access: boolean;
  upgrade_url?: string;
  ideas: Idea[];
  generating: boolean;
  last_batch_at: string | null;
};

export type RunState = "running" | "needs_approval" | "done" | "failed";

export type RunOutput = {
  call_id: string;
  tool: string;
  app: string;
  id: string;
  url: string;
  preview: string;
};

export type RunApproval = {
  call_id: string;
  tool: string;
  args: unknown;
  grantable: boolean;
};

export type IdeaRun = {
  id: string;
  state: RunState;
  summary: string;
  outputs: RunOutput[] | null;
  pending_approvals: RunApproval[] | null;
  continue_url: string;
};

export const IDEA_KINDS = [
  "mail",
  "task",
  "event",
  "follow_up",
  "research",
  "other",
] as const;

export function ideaKind(kind: string): (typeof IDEA_KINDS)[number] {
  return (IDEA_KINDS as readonly string[]).includes(kind)
    ? (kind as (typeof IDEA_KINDS)[number])
    : "other";
}

export function isSafeHref(value: string | undefined | null): value is string {
  if (!value || !value.startsWith("https://")) return false;
  for (const char of value) {
    const code = char.charCodeAt(0);
    if (code <= 0x20 || code === 0x7f || char === "\\") return false;
  }
  try {
    return new URL(value).hostname !== "";
  } catch {
    return false;
  }
}

export function isNewIdea(idea: Idea, now = Date.now()): boolean {
  if (idea.status !== "new") return false;
  const expires = Date.parse(idea.expires_at);
  return Number.isNaN(expires) || expires > now;
}

export function countNewIdeas(
  ideas: readonly Idea[],
  now = Date.now(),
): number {
  return ideas.filter((idea) => isNewIdea(idea, now)).length;
}

export function isRunSettled(state: RunState | undefined): boolean {
  return state === "done" || state === "failed";
}

export type IdeasView = "loading" | "error" | "busy" | "empty" | "list";

export function ideasScope(
  workspaceId: string | undefined,
  workspacesSettled: boolean,
): { ready: boolean; workspaceId: string | undefined } {
  return { ready: !!workspaceId || workspacesSettled, workspaceId };
}

export function ideasView(input: {
  count: number;
  loading: boolean;
  generating: boolean;
  error: string | null;
}): IdeasView {
  if (input.loading) return "loading";
  if (input.count > 0) return "list";
  if (input.error) return "error";
  return input.generating ? "busy" : "empty";
}

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["day", 86_400_000],
  ["hour", 3_600_000],
  ["minute", 60_000],
];

export function relativeTime(
  value: string | null | undefined,
  locale: string,
  now = Date.now(),
): string {
  if (!value) return "";
  const then = new Date(value).getTime();
  if (Number.isNaN(then)) return "";
  const diff = then - now;
  const [unit, size] = UNITS.find(([, ms]) => Math.abs(diff) >= ms) ?? [
    "second",
    1000,
  ];
  return new Intl.RelativeTimeFormat(locale, { numeric: "auto" }).format(
    unit === "second" ? 0 : Math.round(diff / size),
    unit,
  );
}

export function formatArgs(args: unknown): string {
  if (args === undefined || args === null) return "";
  if (typeof args === "string") return args;
  try {
    return JSON.stringify(args, null, 2);
  } catch {
    return String(args);
  }
}

const RUNS_KEY = "opus.ideas.runs";

export function loadRunIds(): Record<string, string> {
  try {
    const parsed = JSON.parse(sessionStorage.getItem(RUNS_KEY) ?? "{}");
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

export function saveRunIds(runs: Record<string, string>) {
  try {
    sessionStorage.setItem(RUNS_KEY, JSON.stringify(runs));
  } catch {
    return;
  }
}
