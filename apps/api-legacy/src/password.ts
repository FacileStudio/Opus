import type { Algorithm } from "@node-rs/argon2";
import { hash as argon2Hash, verify as argon2Verify } from "@node-rs/argon2";
import bcrypt from "bcrypt";

/**
 * Password hashing for email/password accounts.
 *
 * New passwords are Argon2id. The parameters match the Go apps' `authcrypto`
 * so the suite has one answer to "how are passwords hashed here": 64 MiB of
 * memory, 3 passes, 2 lanes.
 *
 * `Algorithm` is declared as an ambient const enum, which `isolatedModules`
 * forbids reading at runtime, so the member is written as its value: 2 is
 * Argon2id. The test asserts the output starts with `$argon2id$`, which is
 * what keeps that number honest.
 */
const ARGON2_OPTIONS = {
  algorithm: 2 as Algorithm,
  memoryCost: 64 * 1024,
  timeCost: 3,
  parallelism: 2,
} as const;

/**
 * bcrypt's encoded form is self-identifying — `$2a$`, `$2b$` or `$2y$` — which
 * is what makes this migration need no schema change and no flag column.
 */
const BCRYPT_HASH = /^\$2[aby]?\$/;

export function hashPassword(password: string): Promise<string> {
  return argon2Hash(password, ARGON2_OPTIONS);
}

/**
 * Verifies against whichever scheme produced the stored hash.
 *
 * Every account created before this change holds a bcrypt hash, and a bcrypt
 * hash cannot be converted without the plaintext. Rejecting them would lock
 * out every existing user, so they keep verifying through bcrypt and move to
 * Argon2id the next time the password is set.
 */
export async function verifyPassword(
  hash: string,
  password: string,
): Promise<boolean> {
  if (BCRYPT_HASH.test(hash)) {
    return bcrypt.compare(password, hash);
  }

  // argon2Verify throws on a malformed hash rather than returning false. A
  // stored value we cannot parse is a failed login, not a 500 on the login
  // route.
  try {
    return await argon2Verify(hash, password);
  } catch {
    return false;
  }
}

/**
 * Whether a stored hash is still on the old scheme. Nothing calls this on the
 * login path — better-auth's verify hook receives only the hash and the
 * password, with no way to write a replacement back — but it is what an
 * offline re-hash or an audit query needs, and it keeps the prefix knowledge
 * in one place.
 */
export function isLegacyHash(hash: string): boolean {
  return BCRYPT_HASH.test(hash);
}
