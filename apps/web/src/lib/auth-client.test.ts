import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "@/fetchers/get-api-url";
import {
  apiErrorMessage,
  apiRequest,
  authClient,
  type User,
} from "./auth-client";

const someone = { id: "1", name: "A", email: "a@facile.studio" } as User;

describe("a revoked session", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("ends the signed-in session once on a 401 from any API call", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(null, { status: 401 })),
    );
    const lost = vi.fn();
    const stop = authClient.onSessionLost(lost);
    authClient.__setSession(someone, false);

    await apiFetch("/api/nook-pool/1");
    await apiFetch("/api/nook-pool/1");
    stop();

    expect(lost).toHaveBeenCalledTimes(1);
  });
});

describe("an API error body", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("gives the nested message as a string, never an object", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              error: { code: "forbidden", message: "Sign in again" },
            }),
            { status: 403 },
          ),
      ),
    );

    const result = await apiRequest("/mcp/pending");

    expect(result.error).toEqual({
      message: "Sign in again",
      status: 403,
      code: "forbidden",
    });
  });

  it("falls back on a flat message, then on the fallback", () => {
    expect(apiErrorMessage({ message: "flat" }, "x")).toBe("flat");
    expect(apiErrorMessage({ error: { code: "c" } }, "x")).toBe("x");
    expect(apiErrorMessage({ error: "plain" }, "x")).toBe("x");
    expect(apiErrorMessage(null, "x")).toBe("x");
  });

  it("sends no bearer token from localStorage", async () => {
    const fetchMock = vi.fn(
      async (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Response(null, { status: 204 }),
    );
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("opus.access_token", "stale");

    await apiRequest("/auth/me");

    const init = fetchMock.mock.calls[0]?.[1];
    expect(new Headers(init?.headers).has("Authorization")).toBe(false);
    localStorage.removeItem("opus.access_token");
  });
});
