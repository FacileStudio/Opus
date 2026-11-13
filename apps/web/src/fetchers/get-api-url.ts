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

/**
 * fetch for the Opus API. A 401 means the session is gone, for example revoked
 * by Registre's back-channel logout, so it is announced to the auth client,
 * which drops the signed-in user and their cached data.
 */
export async function apiFetch(input: string, init?: RequestInit) {
  const response = await fetch(input, init);
  if (response.status === 401) {
    window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
  }
  return response;
}
