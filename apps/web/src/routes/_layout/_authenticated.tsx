import { createFileRoute, redirect } from "@tanstack/react-router";
import { authClient } from "@/lib/auth-client";
import { followRegistreSession } from "@/lib/follow-registre";
import { rememberReturn } from "@/lib/pending";

/**
 * Guards every signed-in page. Without a session the browser goes straight to
 * the SSO login, so a live Registre session signs the person back in with no
 * click. When that already failed once, the sign-in page says so instead.
 */
export const Route = createFileRoute("/_layout/_authenticated")({
  beforeLoad: async ({ location, preload }) => {
    const { data: session } = await authClient.getSession();
    if (!session) {
      if (!preload && followRegistreSession(location.href)) {
        await new Promise<never>(() => {});
      }
      rememberReturn(location.href);
      throw redirect({
        to: "/auth/sign-in",
        search: {
          redirect: location.pathname,
          ended: true,
        },
      });
    }
    return { session };
  },
});
