import { afterEach, describe, expect, it, vi } from "vitest";
import { getApiUrl } from "./get-api-url";

afterEach(() => {
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
