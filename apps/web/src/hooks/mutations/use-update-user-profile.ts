import { useMutation } from "@tanstack/react-query";
import { authClient } from "@/lib/auth-client";

type UpdateUserProfileRequest = {
  name?: string;
  locale?: string;
  color?: string;
};

function useUpdateUserProfile() {
  return useMutation({
    mutationFn: async ({ name, locale, color }: UpdateUserProfileRequest) => {
      const { data, error } = await authClient.updateUser({
        name,
        locale,
        color,
      });

      if (error) {
        throw new Error(error.message || "Failed to update user profile");
      }

      return data;
    },
  });
}

export default useUpdateUserProfile;
