/// <reference types="vite/types/importMeta.d.ts" />

import { hc } from "hono/client";
import { resolveApiBaseUrl } from "./api-url";
import { restringify } from "./stringify-ids";
import type { AppType } from "./types";

const apiUrl = resolveApiBaseUrl(import.meta.env.VITE_API_URL);

/** Dispatched on window whenever the API answers 401: the session is gone. */
export const UNAUTHORIZED_EVENT = "opus:unauthorized";

// Generate once per tab load
export const windowId = Math.random().toString(36).substring(2, 11);

export const client = hc<AppType>(apiUrl, {
  fetch: (input: RequestInfo | URL, init?: RequestInit) => {
    const method = (init?.method || "GET").toUpperCase();
    const headers = new Headers(init?.headers);
    headers.set("Content-Type", "application/json");
    headers.set("X-Opus-Window-Id", windowId);
    // porte refuses a cookie-authenticated mutating request without X-Facile-CSRF.
    if (method !== "GET" && method !== "HEAD") {
      headers.set("X-Facile-CSRF", "1");
    }
    return fetch(input, {
      ...init,
      headers,
      credentials: "include",
    })
      .then(async (response) => {
        if (response.status === 401) {
          window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
        }
        if (!response.headers.get("content-type")?.includes("json")) {
          return response;
        }
        const body = await response.text();
        if (!body) {
          return response;
        }
        const responseHeaders = new Headers(response.headers);
        responseHeaders.delete("content-length");
        responseHeaders.delete("content-encoding");
        return new Response(restringify(body), {
          status: response.status,
          statusText: response.statusText,
          headers: responseHeaders,
        });
      })
      .catch((error) => {
        if (error instanceof TypeError && error.message.includes("fetch")) {
          throw new Error(
            `Failed to connect to API server at ${apiUrl}. This might be due to CORS configuration issues or the server not running. Please check your environment variables and server status.`,
          );
        }
        throw error;
      });
  },
});
