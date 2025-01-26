import type { LoggedInUser } from "@/types/user";
import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/router-devtools";

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient;
  user: LoggedInUser | null | undefined;
}>()({
  component: RootComponent,
});

function RootComponent() {
  return (
    <>
      <div className="flex w-full h-screen flex-col md:flex-row bg-zinc-50 dark:bg-zinc-950 scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-zinc-900">
        <Outlet />
      </div>
      <ReactQueryDevtools buttonPosition="top-right" />
      <TanStackRouterDevtools position="bottom-right" />
    </>
  );
}

export default RootComponent;
