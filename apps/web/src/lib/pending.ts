export type PendingState = "ready" | "changed" | "missing";

export type PendingTarget = {
  type: "task" | "project";
  id: number;
  title: string;
  ref?: string;
  project_name?: string;
  task_count?: number;
  url: string;
};

export type PendingAction = {
  id: string;
  kind: string;
  client_id: string;
  summary: string;
  created_at: string;
  expires_at: string;
  state: PendingState;
  target?: PendingTarget;
};

export const RETURN_KEY = "opus.return";

export function clientName(clientId: string): string {
  try {
    return new URL(clientId).hostname || clientId;
  } catch {
    return clientId;
  }
}

/**
 * An in-app path worth coming back to after sign-in: same origin by
 * construction, and never a sign-in page, which would only sign in again.
 */
export function isReturnPath(href: string): boolean {
  return (
    href.startsWith("/") &&
    !href.startsWith("//") &&
    !href.includes("\\") &&
    !/^\/(auth|login)(\/|\?|#|$)/.test(href)
  );
}

export function rememberReturn(href: string) {
  if (!isReturnPath(href)) return;
  try {
    sessionStorage.setItem(RETURN_KEY, href);
  } catch {
    // storage unavailable: the person opens the page again after sign-in
  }
}

export function takeReturn(): string | null {
  try {
    const href = sessionStorage.getItem(RETURN_KEY);
    sessionStorage.removeItem(RETURN_KEY);
    return href && isReturnPath(href) ? href : null;
  } catch {
    return null;
  }
}
