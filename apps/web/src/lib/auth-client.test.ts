import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "@/fetchers/get-api-url";
import { authClient, type User } from "./auth-client";

const someone = { id: "1", name: "A", email: "a@facile.studio" } as User;

describe("a revoked session", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("ends the signed-in session once on a 401 from any API call", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(null, { status: 401 })),
    );
    const lost = vi.fn();
    const stop = authClient.onSessionLost(lost);
    authClient.__setSession(someone, false);

    await apiFetch("/api/nook-pool/1");
    await apiFetch("/api/nook-pool/1");
    stop();

    expect(lost).toHaveBeenCalledTimes(1);
  });
});
