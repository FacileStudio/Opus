import { describe, expect, it } from "vitest";
import {
  clientName,
  isPendingPath,
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

  it("only returns to the pending page", () => {
    expect(isPendingPath("/pending?action=abc")).toBe(true);
    expect(isPendingPath("/pending")).toBe(true);
    expect(isPendingPath("//evil.com/pending")).toBe(false);
    expect(isPendingPath("/pendingx")).toBe(false);
  });

  it("keeps the confirm link across sign-in, once", () => {
    rememberReturn("/pending?action=abc");
    expect(takeReturn()).toBe("/pending?action=abc");
    expect(takeReturn()).toBeNull();
    rememberReturn("https://evil.com");
    expect(takeReturn()).toBeNull();
  });
});
