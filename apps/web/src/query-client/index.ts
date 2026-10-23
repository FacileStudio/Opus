import { MutationCache, QueryClient } from "@tanstack/react-query";

const queryClient: QueryClient = new QueryClient({
  mutationCache: new MutationCache({
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["my-tasks"] });
    },
  }),
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      refetchOnMount: true,
      retry: (failureCount, error) => {
        if (error instanceof Error) {
          if (
            error.message.includes("Failed to fetch") ||
            error.message.includes("NetworkError") ||
            error.message.includes("CORS")
          ) {
            return false;
          }
        }
        return failureCount < 2;
      },
    },
    mutations: {
      retry: false,
    },
  },
});

export default queryClient;
