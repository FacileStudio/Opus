import i18n from "i18next";
import { toFacileLocale, withUiLocales } from "@/components/auth/facile-locale";
import { getApiUrl } from "@/fetchers/get-api-url";
import queryClient from "@/query-client";
import useBacklogBulkSelectionStore from "@/store/backlog-bulk-selection";
import useBulkSelectionStore from "@/store/bulk-selection";
import useProjectStore from "@/store/project";
import { rememberReturn } from "./pending";

export const SSO_ATTEMPT_KEY = "opus.sso-attempt";

const ATTEMPT_WINDOW_MS = 5 * 60 * 1000;

let leaving = false;

/** Whether this tab was sent to the SSO login moments ago and has not come back signed in. */
export function ssoAttemptPending(now = Date.now()): boolean {
  try {
    const last = Number(sessionStorage.getItem(SSO_ATTEMPT_KEY));
    return Boolean(last) && now - last < ATTEMPT_WINDOW_MS;
  } catch {
    return false;
  }
}

/**
 * Claims the one silent sign-in a lost session gets. It answers false when
 * the previous attempt came back still signed out, and consumes that marker,
 * so the caller shows the sign-in page instead of bouncing through Registre
 * forever. Without storage there is no loop guard, so it answers false too.
 */
export function claimSsoAttempt(now = Date.now()): boolean {
  if (ssoAttemptPending(now)) {
    clearSsoAttempt();
    return false;
  }
  try {
    sessionStorage.setItem(SSO_ATTEMPT_KEY, String(now));
    return true;
  } catch {
    return false;
  }
}

/** Forgets the attempt once a session is confirmed, so the next loss gets its own. */
export function clearSsoAttempt() {
  try {
    sessionStorage.removeItem(SSO_ATTEMPT_KEY);
  } catch {
    return;
  }
}

/**
 * Sends the browser to porte's SSO login so Registre's live session, whoever
 * it belongs to now, signs the person straight back in, then returns them to
 * returnTo. Answers false without navigating when the attempt is refused. A
 * second call while the page is already leaving answers true and leaves the
 * marker alone, so the session-lost listener and the route guard firing on
 * the same 401 cannot spend the one attempt between them.
 */
export function followRegistreSession(returnTo: string): boolean {
  if (leaving) return true;
  if (!claimSsoAttempt()) return false;
  leaving = true;
  rememberReturn(returnTo);
  window.location.assign(
    withUiLocales(getApiUrl("/auth/oidc"), toFacileLocale(i18n.language)),
  );
  return true;
}

/** Drops everything cached for the person whose session just ended. */
export function forgetUserData() {
  queryClient.clear();
  useProjectStore.getState().setProject(undefined);
  useBulkSelectionStore.getState().clearSelection();
  useBacklogBulkSelectionStore.getState().clearSelection();
}
