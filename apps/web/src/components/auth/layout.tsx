import { useTranslation } from "react-i18next";
import { Logo } from "@/components/common/logo";

type AuthLayoutProps = {
  children: React.ReactNode;
  title: string;
  subtitle?: string;
};

export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  const { t } = useTranslation();
  return (
    <div className="flex h-svh w-full overflow-y-auto bg-background text-foreground">
      <div className="hidden w-1/2 flex-col justify-between bg-black px-12 py-10 lg:flex">
        <Logo to="/" wordmark className="text-white" />
        <div>
          <h2 className="font-heading text-4xl font-semibold leading-tight tracking-tight text-white">
            {t("auth:layout.headline")}
            <br />
            {t("auth:layout.headlineEnd")}
          </h2>
          <p className="mt-4 max-w-xs text-sm leading-relaxed text-white/50">
            {t("auth:signIn.tagline")}
          </p>
        </div>
        <p className="text-xs text-white/30">
          {t("landing:copyright", { year: new Date().getFullYear() })}
        </p>
      </div>
      <div className="flex w-full flex-col items-center justify-center bg-background px-8 py-12 lg:w-1/2">
        <div className="w-full max-w-sm">
          <Logo to="/" wordmark className="mb-8 lg:hidden" />
          <div className="mb-8">
            <h1 className="font-heading text-2xl font-semibold tracking-tight text-foreground">
              {title}
            </h1>
            {subtitle ? (
              <p className="mt-1.5 text-sm text-muted-foreground">{subtitle}</p>
            ) : null}
          </div>
          {children}
        </div>
      </div>
    </div>
  );
}
