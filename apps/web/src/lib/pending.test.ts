import { describe, expect, it } from "vitest";
import {
  clientName,
  isReturnPath,
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
});
