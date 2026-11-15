import { UNAUTHORIZED_EVENT } from "@opus/libs";

export function getApiUrl(path: string) {
  const trimmedBase = (
    import.meta.env.VITE_API_URL || "http://localhost:1337"
  ).replace(/\/+$/, "");
  const apiUrl = trimmedBase.endsWith("/api")
    ? trimmedBase
    : `${trimmedBase}/api`;
  const normalizedPath = `/${path.replace(/^\/+/, "")}`;

  return `${apiUrl}${normalizedPath}`;
}

const CSRF_HEADER = "X-Facile-CSRF";
const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

/**
 * fetch for the Opus API. A cookie-authenticated request that is not GET, HEAD
 * or OPTIONS must carry X-Facile-CSRF, so it is added here unless the caller
 * set it. A 401 means the session is gone, for example revoked by Registre's
 * back-channel logout, so it is announced to the auth client, which drops the
 * signed-in user and their cached data.
 */
export async function apiFetch(input: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers);
  const method = (init.method ?? "GET").toUpperCase();
  if (!SAFE_METHODS.has(method) && !headers.has(CSRF_HEADER)) {
    headers.set(CSRF_HEADER, "1");
  }
  const response = await fetch(input, { ...init, headers });
  if (response.status === 401) {
    window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
  }
  return response;
}
