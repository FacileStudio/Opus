import { createContext, type PropsWithChildren, useEffect } from "react";
import { authClient } from "@/lib/auth-client";
import type { User } from "@/types/user";
import { LoadingSkeleton } from "../../ui/loading-skeleton";

const { useSession } = authClient;

export const AuthContext = createContext<{
  user: User | null | undefined;
  isLoading: boolean;
}>({
  user: undefined,
  isLoading: true,
});

function AuthProvider({ children }: PropsWithChildren) {
  const { data, isPending } = useSession();

  useEffect(() => {
    let active = true;

    const refresh = () => {
      authClient.getSession().then(({ data: session }) => {
        if (!active) return;
        authClient.__setSession(
          (session?.user as User | null | undefined) ?? null,
          false,
        );
      });
    };

    refresh();
    // An idle tab can outlive its session cookie: re-verify when the tab
    // regains focus so a dead session does not keep rendering as logged in.
    window.addEventListener("focus", refresh);
    const onVisibility = () => {
      if (document.visibilityState === "visible") {
        refresh();
      }
    };
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      active = false;
      window.removeEventListener("focus", refresh);
      document.removeEventListener("visibilitychange", onVisibility);
    };
  }, []);

  if (isPending) {
    return <LoadingSkeleton />;
  }

  return (
    <AuthContext.Provider
      value={{
        user: (data?.user as User | null | undefined) ?? null,
        isLoading: isPending,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export default AuthProvider;
