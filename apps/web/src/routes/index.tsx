import {
  ChatRound,
  ClockCircle,
  Layers,
  Server,
  Settings,
  UsersGroupTwoRounded,
} from "@solar-icons/react";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { SignInWithFacile } from "@/components/auth/sign-in-with-facile";
import { Logo } from "@/components/common/logo";
import PageTitle from "@/components/page-title";
import { Card, CardContent } from "@/components/ui/card";
import { authClient } from "@/lib/auth-client";
import { ssoAttemptPending } from "@/lib/follow-registre";

type Feature = {
  icon: typeof Layers;
  key: string;
};

const features: Feature[] = [
  { icon: Layers, key: "views" },
  { icon: ChatRound, key: "realtime" },
  { icon: Settings, key: "rules" },
  { icon: ClockCircle, key: "time" },
  { icon: UsersGroupTwoRounded, key: "workspaces" },
  { icon: Server, key: "selfHosted" },
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
    if (ssoAttemptPending()) {
      throw redirect({ to: "/auth/sign-in", search: { ended: true } });
    }
  },
});

function Landing() {
  const { t } = useTranslation("landing");

  return (
    <div className="flex h-svh w-full flex-col overflow-y-auto bg-background text-foreground">
      <PageTitle title={t("pageTitle")} hideAppName />

      <header>
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-4">
          <Logo to="/" wordmark />
          <SignInWithFacile size="sm" />
        </div>
      </header>

      <main className="flex-1">
        <section className="mx-auto w-full max-w-5xl px-6 py-16 text-center sm:py-24">
          <h1 className="mx-auto max-w-3xl text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
            {t("heroTitle")}
          </h1>
          <p className="mx-auto mt-4 max-w-xl text-base text-muted-foreground">
            {t("heroBody")}
          </p>
          <div className="mt-8 flex items-center justify-center">
            <SignInWithFacile size="lg" />
          </div>
        </section>

        <section className="mx-auto w-full max-w-5xl px-6 pb-20">
          <h2 className="text-2xl font-semibold tracking-tight">
            {t("sectionTitle")}
          </h2>
          <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map(({ icon: Icon, key }) => (
              <Card key={key}>
                <CardContent className="p-5">
                  <Icon className="size-6" aria-hidden="true" />
                  <h3 className="mt-4 text-base font-medium">
                    {t(`features.${key}.title`)}
                  </h3>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {t(`features.${key}.body`)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      </main>

      <footer>
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-6">
          <p className="text-xs text-muted-foreground">
            {t("copyright", { year: new Date().getFullYear() })}
          </p>
          <p className="text-xs text-muted-foreground">{t("footerTagline")}</p>
        </div>
      </footer>
    </div>
  );
}
