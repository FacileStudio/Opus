import type { SignUpFormValues } from "@/components/auth/sign-up-form";
import signUp from "@/fetchers/user/sign-up";
import { useMutation } from "@tanstack/react-query";

function useSignUp() {
  return useMutation({
    mutationFn: ({ email, password, name }: SignUpFormValues) =>
      signUp({
        email,
        password,
        name,
      }),
  });
}

export default useSignUp;
