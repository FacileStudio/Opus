import me from "@/fetchers/user/me";
import { useQuery } from "@tanstack/react-query";

function useGetMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: () => me(),
    retry: 0,
  });
}

export default useGetMe;
