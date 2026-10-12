import {
  ChatRound,
  ClockCircle,
  Layers,
  Server,
  Settings,
  UsersGroupTwoRounded,
} from "@solar-icons/react";
import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Logo } from "@/components/common/logo";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import useGetConfig from "@/hooks/queries/config/use-get-config";
import { authClient } from "@/lib/auth-client";

type Feature = {
  icon: typeof Layers;
  title: string;
  body: string;
};

const features: Feature[] = [
  {
    icon: Layers,
    title: "Four views, one backlog",
    body: "Board, list, gantt and backlog over the same tasks. Switch per project without moving a thing.",
  },
  {
    icon: ChatRound,
    title: "Realtime collaboration",
    body: "Comments, labels and task moves land in every open tab the moment somebody makes them.",
  },
  {
    icon: Settings,
    title: "Workflow rules",
    body: "Status transitions, assignments and due-date reminders run on the server instead of in your head.",
  },
  {
    icon: ClockCircle,
    title: "Time tracking",
    body: "Log time against tasks and read it back per project when the review starts.",
  },
  {
    icon: UsersGroupTwoRounded,
    title: "Workspaces and roles",
    body: "Invite a team, keep projects private or shared, and let owner, admin and member decide who edits what.",
  },
  {
    icon: Server,
    title: "Self-hosted by default",
    body: "One container, your Postgres, your storage. Nothing you plan leaves the server it runs on.",
  },
];

export const Route = createFileRoute("/")({
  component: Landing,
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
  },
});

function Landing() {
  const { data } = useGetConfig();
  const ssoOnly = Boolean(data?.ssoOnly);
  const cta = ssoOnly
    ? ({ to: "/auth/sign-in", label: "Sign in with Facile" } as const)
    : ({ to: "/auth/sign-up", label: "Get started" } as const);

  return (
    <div className="flex h-svh w-full flex-col overflow-y-auto bg-background text-foreground">
      <PageTitle title="Self-hosted project management" />

      <header className="border-b border-border">
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-4">
          <Logo to="/" wordmark />
          <div className="flex items-center gap-2">
            <Link
              to="/auth/sign-in"
              className="inline-flex h-9 items-center justify-center rounded-lg px-4 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              Log in
            </Link>
            <Link
              to={cta.to}
              className="inline-flex h-9 items-center justify-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
            >
              {cta.label}
            </Link>
          </div>
        </div>
      </header>

      <main className="flex-1">
        <section className="mx-auto w-full max-w-5xl px-6 py-16 text-center sm:py-24">
          <h1 className="mx-auto max-w-3xl text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
            Project management you host yourself
          </h1>
          <p className="mx-auto mt-4 max-w-xl text-base text-muted-foreground">
            Opus keeps tasks, boards and timelines on your own server. Your team
            plans where your data already lives.
          </p>
          <div className="mt-8 flex items-center justify-center gap-3">
            <Button render={<Link to={cta.to} />} size="lg">
              {cta.label}
            </Button>
            <Button
              render={<Link to="/auth/sign-in" />}
              variant="ghost"
              size="lg"
            >
              Log in
            </Button>
          </div>
        </section>

        <section className="mx-auto w-full max-w-5xl px-6 pb-20">
          <h2 className="text-2xl font-semibold tracking-tight">
            Everything a project needs, nothing it does not
          </h2>
          <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map(({ icon: Icon, title, body }) => (
              <Card key={title}>
                <CardContent className="p-5">
                  <Icon className="size-6" aria-hidden="true" />
                  <h3 className="mt-4 text-base font-medium">{title}</h3>
                  <p className="mt-1 text-sm text-muted-foreground">{body}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      </main>

      <footer className="border-t border-border">
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-6">
          <p className="text-xs text-muted-foreground">
            © {new Date().getFullYear()} Opus by Facile.
          </p>
          <p className="text-xs text-muted-foreground">
            Self-hosted project management.
          </p>
        </div>
      </footer>
    </div>
  );
}
