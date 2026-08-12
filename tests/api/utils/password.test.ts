import bcrypt from "bcrypt";
import { describe, expect, it } from "vitest";
import {
  hashPassword,
  isLegacyHash,
  verifyPassword,
} from "../../../apps/api-legacy/src/password";

const PASSWORD = "correct horse battery staple";

describe("password hashing", () => {
  it("hashes new passwords with argon2id", async () => {
    const hash = await hashPassword(PASSWORD);
    expect(hash.startsWith("$argon2id$")).toBe(true);
    expect(isLegacyHash(hash)).toBe(false);
  });

  it("verifies a password it just hashed", async () => {
    const hash = await hashPassword(PASSWORD);
    await expect(verifyPassword(hash, PASSWORD)).resolves.toBe(true);
    await expect(verifyPassword(hash, "wrong")).resolves.toBe(false);
  });

  it("still verifies bcrypt hashes, so existing accounts are not locked out", async () => {
    // Every account that existed before the switch holds one of these, and a
    // bcrypt hash cannot be converted without the plaintext. If this test ever
    // fails, every pre-existing user has lost their password.
    const legacy = await bcrypt.hash(PASSWORD, 10);
    expect(isLegacyHash(legacy)).toBe(true);
    await expect(verifyPassword(legacy, PASSWORD)).resolves.toBe(true);
    await expect(verifyPassword(legacy, "wrong")).resolves.toBe(false);
  });

  it("treats an unparseable stored hash as a failed login, not an exception", async () => {
    await expect(verifyPassword("not-a-hash", PASSWORD)).resolves.toBe(false);
    await expect(verifyPassword("", PASSWORD)).resolves.toBe(false);
  });

  it("salts, so the same password hashes differently twice", async () => {
    const [a, b] = await Promise.all([
      hashPassword(PASSWORD),
      hashPassword(PASSWORD),
    ]);
    expect(a).not.toBe(b);
  });
});
