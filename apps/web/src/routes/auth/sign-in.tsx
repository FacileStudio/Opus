import { createFileRoute, useSearch } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { z } from "zod/v4";
import { AuthScreen } from "@/components/auth/auth-screen";
import PageTitle from "@/components/page-title";

const signInSearchSchema = z.object({
  invitationId: z.string().optional(),
  email: z.string().optional(),
  redirect: z.string().optional(),
  error: z.string().optional(),
  ended: z.boolean().optional(),
});

export const Route = createFileRoute("/auth/sign-in")({
  component: SignIn,
  validateSearch: signInSearchSchema,
});

function SignIn() {
  const { t } = useTranslation();
  const search = useSearch({ from: "/auth/sign-in" });

  const notice = search.error
    ? `${t("auth:signIn.oidcError")}: ${search.error}`
    : search.ended
      ? t("auth:signIn.sessionEnded")
      : search.invitationId
        ? t("auth:signIn.invitationAlert")
        : null;

  return (
    <>
      <PageTitle title={t("auth:signIn.pageTitle")} />
      <AuthScreen
        notice={notice}
        noticeTone={search.error || search.ended ? "error" : "info"}
      />
    </>
  );
}
