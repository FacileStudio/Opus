import {
  createFileRoute,
  useNavigate,
  useSearch,
} from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { z } from "zod/v4";
import PageTitle from "@/components/page-title";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import useGetConfig from "@/hooks/queries/config/use-get-config";
import { AuthLayout } from "../../components/auth/layout";
import { SignInForm } from "../../components/auth/sign-in-form";
import { SignInFormSkeleton } from "../../components/auth/sign-in-form-skeleton";
import { AuthToggle } from "../../components/auth/toggle";

const signInSearchSchema = z.object({
  invitationId: z.string().optional(),
  email: z.string().optional(),
  redirect: z.string().optional(),
  error: z.string().optional(),
});

export const Route = createFileRoute("/auth/sign-in")({
  component: SignIn,
  validateSearch: signInSearchSchema,
});

function getApiBase() {
  return (import.meta.env.VITE_API_URL || "http://localhost:1337").replace(
    /\/+$/,
    "",
  );
}

function SignIn() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const search = useSearch({ from: "/auth/sign-in" });
  const { data, isLoading: isConfigLoading } = useGetConfig();
  const config = (data ?? {}) as
    | {
        ssoOnly?: boolean | null;
        hasCustomOAuth?: boolean | null;
        disableRegistration?: boolean | null;
        disablePasswordRegistration?: boolean | null;
      }
    | undefined;

  const invitationId = search.invitationId;
  const defaultEmail = search.email;
  const oidcError = search.error;

  const getSafeRedirectPath = () => {
    const redirectPath = search.redirect;
    if (redirectPath?.startsWith("/") && !redirectPath.includes("//")) {
      return redirectPath;
    }
    return undefined;
  };

  const handleSso = () => {
    window.location.href = `${getApiBase()}/api/auth/oidc`;
  };

  const handleSignInSuccess = () => {
    const redirectPath = getSafeRedirectPath();
    if (redirectPath) {
      navigate({ to: redirectPath });
    } else if (invitationId) {
      navigate({ to: `/invitation/accept/${invitationId}` });
    } else {
      navigate({ to: "/dashboard" });
    }
  };

  if (isConfigLoading) {
    return (
      <>
        <PageTitle title={t("auth:signIn.pageTitle")} />
        <AuthLayout
          title={t("auth:signIn.title")}
          subtitle={t("auth:signIn.subtitle")}
        >
          <SignInFormSkeleton />
        </AuthLayout>
      </>
    );
  }

  const ssoOnly = config?.ssoOnly;
  const showSso = Boolean(ssoOnly || config?.hasCustomOAuth);

  return (
    <>
      <PageTitle title={t("auth:signIn.pageTitle")} />
      <AuthLayout
        title={t("auth:signIn.title")}
        subtitle={
          invitationId
            ? t("auth:signIn.invitationSubtitle")
            : t("auth:signIn.subtitle")
        }
      >
        <div className="mt-6">
          {oidcError && (
            <Alert variant="error" className="mb-4">
              <AlertDescription>
                {t("auth:signIn.oidcError")}: {oidcError}
              </AlertDescription>
            </Alert>
          )}

          {invitationId && (
            <Alert className="mb-4">
              <AlertDescription>
                {t("auth:signIn.invitationAlert")}
              </AlertDescription>
            </Alert>
          )}

          {showSso && (
            <>
              <Button variant="outline" onClick={handleSso} className="w-full">
                {t("auth:signIn.continueWithOidc")}
              </Button>

              {!ssoOnly && (
                <div className="relative my-6">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-border" />
                  </div>
                  <div className="relative flex justify-center text-sm">
                    <span className="px-2 bg-card text-muted-foreground">
                      {t("auth:forms.or")}
                    </span>
                  </div>
                </div>
              )}
            </>
          )}

          {!ssoOnly && (
            <SignInForm
              defaultEmail={defaultEmail}
              onSuccess={handleSignInSuccess}
            />
          )}

          {ssoOnly ? (
            <div className="text-center pt-4">
              <p className="text-sm text-muted-foreground">
                {t("auth:signIn.ssoOnlyMessage")}
              </p>
            </div>
          ) : config?.disableRegistration ||
            config?.disablePasswordRegistration ? (
            <div className="text-center pt-4">
              <p className="text-sm text-muted-foreground">
                {config?.disableRegistration
                  ? t("auth:signIn.registrationDisabled")
                  : t("auth:signIn.passwordRegistrationDisabled")}
              </p>
            </div>
          ) : (
            <AuthToggle
              message={t("auth:signIn.toggleMessage")}
              linkText={t("auth:signIn.toggleLink")}
              linkTo="/auth/sign-up"
            />
          )}
        </div>
      </AuthLayout>
    </>
  );
}
