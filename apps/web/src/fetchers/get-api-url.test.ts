import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch, getApiUrl } from "./get-api-url";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

describe("getApiUrl", () => {
  it("keeps one /api when the base already ends with it", () => {
    vi.stubEnv("VITE_API_URL", "https://opus.facile.studio/api");
    expect(getApiUrl("/auth/oidc")).toBe(
      "https://opus.facile.studio/api/auth/oidc",
    );
  });

  it("appends /api when the base carries none", () => {
    vi.stubEnv("VITE_API_URL", "https://opus.facile.studio");
    expect(getApiUrl("/auth/me")).toBe(
      "https://opus.facile.studio/api/auth/me",
    );
  });

  it("ignores trailing slashes on the base", () => {
    vi.stubEnv("VITE_API_URL", "https://opus.facile.studio/api/");
    expect(getApiUrl("auth/me")).toBe("https://opus.facile.studio/api/auth/me");
  });
});

describe("apiFetch", () => {
  async function sentHeaders(init?: RequestInit) {
    const fetchMock = vi.fn(async () => new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);
    await apiFetch("https://opus.facile.studio/api/x", init);
    const calls = fetchMock.mock.calls as unknown as [string, RequestInit][];
    return new Headers(calls[0][1].headers);
  }

  it.each([
    "POST",
    "PUT",
    "PATCH",
    "DELETE",
    "post",
  ])("adds the CSRF header on %s", async (method) => {
    expect((await sentHeaders({ method })).get("X-Facile-CSRF")).toBe("1");
  });

  it.each([
    undefined,
    "GET",
    "HEAD",
    "OPTIONS",
  ])("leaves %s without the CSRF header", async (method) => {
    expect((await sentHeaders({ method })).has("X-Facile-CSRF")).toBe(false);
  });

  it("keeps a header the caller set and the other headers", async () => {
    const headers = await sentHeaders({
      method: "POST",
      headers: { "X-Facile-CSRF": "custom", Accept: "application/json" },
    });
    expect(headers.get("X-Facile-CSRF")).toBe("custom");
    expect(headers.get("Accept")).toBe("application/json");
  });
});
