import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  claimSsoAttempt,
  clearSsoAttempt,
  followRegistreSession,
  ssoAttemptPending,
} from "./follow-registre";
import { takeReturn } from "./pending";

describe("following the Registre session", () => {
  beforeEach(() => sessionStorage.clear());
  afterEach(() => vi.unstubAllGlobals());

  it("redirects once, then gives up until a session is confirmed", () => {
    expect(claimSsoAttempt(1_000)).toBe(true);
    expect(ssoAttemptPending(2_000)).toBe(true);
    expect(claimSsoAttempt(2_000)).toBe(false);
    expect(ssoAttemptPending(2_000)).toBe(false);
    expect(claimSsoAttempt(3_000)).toBe(true);
    clearSsoAttempt();
    expect(claimSsoAttempt(4_000)).toBe(true);
  });

  it("forgets an attempt that never came back", () => {
    expect(claimSsoAttempt(1)).toBe(true);
    expect(claimSsoAttempt(10 * 60 * 1000)).toBe(true);
  });

  it("sends the browser to the SSO login once per page and keeps the path", () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { ...window.location, assign });
    expect(followRegistreSession("/dashboard/workspace/7")).toBe(true);
    expect(followRegistreSession("/dashboard/workspace/7")).toBe(true);
    expect(assign).toHaveBeenCalledTimes(1);
    expect(assign.mock.calls[0]?.[0]).toMatch(/\/auth\/oidc/);
    expect(ssoAttemptPending()).toBe(true);
    expect(takeReturn()).toBe("/dashboard/workspace/7");
  });
});
