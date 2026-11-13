// @vitest-environment node
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { describe, expect, it } from "vitest";

const script = path.resolve(__dirname, "../../env.sh");

describe("env.sh", () => {
  it("substitutes the two public URLs and never prints or injects other OPUS_ values", () => {
    const dir = mkdtempSync(path.join(tmpdir(), "opus-env-"));
    const bundle = path.join(dir, "app.js");
    writeFileSync(
      bundle,
      'a="OPUS_API_URL";b="OPUS_CLIENT_URL";c="OPUS_SECRET_KEY";',
    );
    const output = execFileSync("sh", [script], {
      env: {
        PATH: process.env.PATH,
        CLIENT_DIR: dir,
        OPUS_API_URL: "https://example.test/api",
        OPUS_CLIENT_URL: "https://example.test",
        OPUS_SECRET_KEY: "top-secret-value",
      },
    }).toString();
    const rewritten = readFileSync(bundle, "utf8");
    expect(rewritten).toBe(
      'a="https://example.test/api";b="https://example.test";c="OPUS_SECRET_KEY";',
    );
    expect(output).not.toContain("top-secret-value");
    expect(output).not.toContain("https://example.test");
  });
});
