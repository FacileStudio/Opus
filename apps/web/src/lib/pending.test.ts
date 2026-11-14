import { describe, expect, it } from "vitest";
import {
  clientName,
  isOwnSessionRefusal,
  isReturnPath,
  OWN_SESSION_REFUSAL,
  rememberReturn,
  takeReturn,
} from "./pending";

describe("pending helpers", () => {
  it("names a client by its host", () => {
    expect(
      clientName("https://claude.ai/oauth/claude-code-client-metadata"),
    ).toBe("claude.ai");
    expect(clientName("bulle")).toBe("bulle");
  });

  it("returns to in-app pages only", () => {
    expect(isReturnPath("/dashboard/workspace/1?tab=a#x")).toBe(true);
    expect(isReturnPath("//evil.com/x")).toBe(false);
    expect(isReturnPath("/\\evil.com")).toBe(false);
    expect(isReturnPath("https://evil.com/")).toBe(false);
    expect(isReturnPath("/auth/sign-in?redirect=/x")).toBe(false);
    expect(isReturnPath("/login?error=x")).toBe(false);
    expect(isReturnPath("/authors")).toBe(true);
  });

  it("keeps the confirm link across sign-in, once", () => {
    rememberReturn("/pending?action=abc");
    expect(takeReturn()).toBe("/pending?action=abc");
    expect(takeReturn()).toBeNull();
    rememberReturn("https://evil.com");
    expect(takeReturn()).toBeNull();
  });

  it("asks to sign in again only for porte's own-session refusal", () => {
    expect(
      isOwnSessionRefusal({
        status: 403,
        code: "permission_denied",
        message: OWN_SESSION_REFUSAL,
      }),
    ).toBe(true);
    expect(
      isOwnSessionRefusal({
        status: 403,
        code: "permission_denied",
        message: "access denied",
      }),
    ).toBe(false);
    expect(
      isOwnSessionRefusal({ status: 403, message: OWN_SESSION_REFUSAL }),
    ).toBe(false);
    expect(isOwnSessionRefusal(null)).toBe(false);
  });
});
