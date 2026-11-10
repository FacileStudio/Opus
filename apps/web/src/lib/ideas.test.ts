import { describe, expect, it } from "vitest";
import {
  countNewIdeas,
  IDEAS_ICON,
  type Idea,
  ideasView,
  isSafeHref,
} from "./ideas";

const idea = (over: Partial<Idea>): Idea => ({
  id: "1",
  kind: "task",
  title: "t",
  why: "w",
  draft: "",
  sources: null,
  status: "new",
  created_at: "2026-11-10T00:00:00Z",
  expires_at: "2026-11-20T00:00:00Z",
  ...over,
});

describe("ideas helpers", () => {
  it("uses the lightbulb icon", () => {
    expect(IDEAS_ICON).toBe("solar:lightbulb-linear");
  });

  it("counts only new, unexpired ideas", () => {
    const now = Date.parse("2026-11-12T00:00:00Z");
    const ideas = [
      idea({}),
      idea({ id: "2", status: "dismissed" }),
      idea({ id: "3", expires_at: "2026-11-11T00:00:00Z" }),
    ];
    expect(countNewIdeas(ideas, now)).toBe(1);
  });

  it("accepts https links only", () => {
    expect(isSafeHref("https://bulle.facile.studio/chat/1")).toBe(true);
    expect(isSafeHref("javascript:alert(1)")).toBe(false);
    expect(isSafeHref("https:evil.com")).toBe(false);
    expect(isSafeHref("https://a b.com")).toBe(false);
    expect(isSafeHref("http://example.com")).toBe(false);
  });

  it("picks the page body", () => {
    const base = { count: 0, loading: false, generating: false, error: null };
    expect(ideasView({ ...base, loading: true })).toBe("loading");
    expect(ideasView({ ...base, error: "x" })).toBe("error");
    expect(ideasView({ ...base, generating: true })).toBe("busy");
    expect(ideasView(base)).toBe("empty");
    expect(ideasView({ ...base, count: 2, error: "x" })).toBe("list");
  });
});
