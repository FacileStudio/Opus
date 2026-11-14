import { createFileRoute, redirect } from "@tanstack/react-router";
import { authClient } from "@/lib/auth-client";
import { ssoAttemptPending } from "@/lib/follow-registre";

export const Route = createFileRoute("/")({
  beforeLoad: async () => {
    let signedIn = false;
    try {
      signedIn = (await authClient.getSession()).data !== null;
    } catch {
      signedIn = false;
    }
    if (signedIn) {
      throw redirect({ to: "/dashboard" });
    }
    throw redirect({
      to: "/auth/sign-in",
      search: ssoAttemptPending() ? { ended: true } : {},
    });
  },
});
