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
    authClient.getSession().then(({ data: session }) => {
      if (!active) return;
      authClient.__setSession(
        (session?.user as User | null | undefined) ?? null,
        false,
      );
    });
    return () => {
      active = false;
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
