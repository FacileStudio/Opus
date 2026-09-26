import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { authClient } from "@/lib/auth-client";

function useSignOut(idpLogoutUrl?: string | null) {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async () => {
      const result = await authClient.signOut();
      if (result.error) {
        throw new Error(result.error.message);
      }
      if (idpLogoutUrl) {
        const redirectUri = `${window.location.origin}/auth/sign-in`;
        const url = new URL(idpLogoutUrl);
        url.searchParams.set("post_logout_redirect_uri", redirectUri);
        window.location.href = url.toString();
      } else {
        navigate({ to: "/auth/sign-in" });
      }
      return result.data;
    },
  });
}

export default useSignOut;
