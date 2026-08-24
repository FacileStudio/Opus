import { createHmac, timingSafeEqual } from "node:crypto";
import type { Context } from "hono";
import { HTTPException } from "hono/http-exception";
import { auth } from "../auth";
import { verifyApiKey } from "./verify-api-key";

const BRIDGE_AUTH_HEADER = "X-Opus-Auth";
const BRIDGE_USER_ID_HEADER = "X-Opus-User-Id";
const BRIDGE_USER_EMAIL_HEADER = "X-Opus-User-Email";

/**
 * The Go proxy stamps proxied requests with the porte-verified identity and
 * an HMAC tag derived from AUTH_BRIDGE_SECRET. When the tag verifies, the
 * legacy API trusts the headers and skips its own better-auth lookup; the
 * browser's porte session means nothing to better-auth, so without the bridge
 * every proxied route would 401. When the tag does not verify the headers are
 * ignored entirely and the request falls through to the normal auth paths,
 * which is what makes a misconfigured bridge fail closed.
 */
function bridgedIdentity(c: Context): { userId: string; email: string } | null {
  const tag = c.req.header(BRIDGE_AUTH_HEADER);
  const userId = c.req.header(BRIDGE_USER_ID_HEADER);
  const email = c.req.header(BRIDGE_USER_EMAIL_HEADER);
  const secret = process.env.AUTH_BRIDGE_SECRET;
  if (!tag || !userId || !secret) {
    return null;
  }
  const expected = createHmac("sha256", secret).update(userId).digest();
  const actual = Buffer.from(tag, "hex");
  if (actual.length !== expected.length) {
    return null;
  }
  if (!timingSafeEqual(actual, expected)) {
    return null;
  }
  return { userId, email: email ?? "" };
}

function applyBridgedIdentity(c: Context): boolean {
  const bridged = bridgedIdentity(c);
  if (!bridged) {
    return false;
  }
  c.set("userId", bridged.userId);
  c.set("userEmail", bridged.email);
  c.set("user", null);
  c.set("session", null);
  return true;
}

async function getSessionFromBearerOnlyHeaders(c: Context) {
  const headers = new Headers(c.req.raw.headers);
  headers.delete("cookie");

  return auth.api.getSession({ headers });
}

function parseBearerToken(authHeader: string | undefined): {
  token: string | null;
  malformed: boolean;
} {
  if (!authHeader) {
    return { token: null, malformed: false };
  }

  if (!authHeader.match(/^Bearer\b/i)) {
    return { token: null, malformed: false };
  }

  const match = authHeader.match(/^Bearer\s+(\S+)$/i);
  if (!match) {
    return { token: null, malformed: true };
  }

  return {
    token: match[1],
    malformed: false,
  };
}

export async function authenticateApiRequest(c: Context): Promise<void> {
  const { token, malformed } = parseBearerToken(c.req.header("Authorization"));
  if (malformed) {
    throw new HTTPException(401, { message: "Unauthorized" });
  }

  if (applyBridgedIdentity(c)) {
    return;
  }

  if (token) {
    const apiKeyResult = await verifyApiKey(token);
    if (apiKeyResult?.valid && apiKeyResult.key) {
      const key = apiKeyResult.key;
      c.set("userId", key.userId);
      c.set("userEmail", "");
      c.set("user", null);
      c.set("session", null);
      c.set("apiKey", {
        id: key.id,
        userId: key.userId,
        enabled: key.enabled,
      });
      return;
    }
    const sessionResult = await getSessionFromBearerOnlyHeaders(c);
    if (sessionResult?.user && sessionResult.session) {
      c.set("user", sessionResult.user);
      c.set("session", sessionResult.session);
      c.set("userId", sessionResult.user.id);
      c.set("userEmail", sessionResult.user.email ?? "");
      return;
    }
    throw new HTTPException(401, { message: "Unauthorized" });
  }

  const sessionResult = await auth.api.getSession({
    headers: c.req.raw.headers,
  });
  c.set("user", sessionResult?.user ?? null);
  c.set("session", sessionResult?.session ?? null);
  c.set("userId", sessionResult?.user?.id ?? "");
  c.set("userEmail", sessionResult?.user?.email ?? "");

  if (!sessionResult?.user) {
    throw new HTTPException(401, { message: "Unauthorized" });
  }
}

export async function resolveAssetBearerOrCookie(c: Context): Promise<{
  userId: string;
  apiKeyId?: string;
}> {
  const { token, malformed } = parseBearerToken(c.req.header("Authorization"));
  if (malformed) {
    throw new HTTPException(401, { message: "Unauthorized" });
  }

  const bridged = bridgedIdentity(c);
  if (bridged) {
    return { userId: bridged.userId };
  }

  if (token) {
    const apiKeyResult = await verifyApiKey(token);
    if (apiKeyResult?.valid && apiKeyResult.key) {
      return {
        userId: apiKeyResult.key.userId,
        apiKeyId: apiKeyResult.key.id,
      };
    }
    const sessionResult = await getSessionFromBearerOnlyHeaders(c);
    if (sessionResult?.user?.id) {
      return { userId: sessionResult.user.id };
    }
    throw new HTTPException(401, { message: "Unauthorized" });
  }

  const sessionResult = await auth.api.getSession({
    headers: c.req.raw.headers,
  });
  if (!sessionResult?.user) {
    throw new HTTPException(401, { message: "Unauthorized" });
  }

  return { userId: sessionResult.user.id };
}
