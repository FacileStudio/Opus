import me from "@/fetchers/user/me";
import { useQuery } from "@tanstack/react-query";

function useGetMe(userId: string | undefined) {
  console.log({ userId });

  return useQuery({
    queryKey: ["me", userId],
    queryFn: () => me(),
    retry: 0,
  });
}

export default useGetMe;
