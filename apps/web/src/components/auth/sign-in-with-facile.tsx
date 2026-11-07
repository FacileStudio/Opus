import { useState } from "react";
import { useTranslation } from "react-i18next";
import { getApiUrl } from "@/fetchers/get-api-url";
import { toFacileLocale, withUiLocales } from "./facile-locale";

export type SignInSize = "sm" | "md" | "lg";

type SignInWithFacileProps = {
  href?: string;
  size?: SignInSize;
  fullWidth?: boolean;
  loading?: boolean;
  className?: string;
};

const SIZE_CLASS: Record<SignInSize, string> = {
  sm: " fc-signin--sm",
  md: "",
  lg: " fc-signin--lg",
};

export function SignInWithFacile({
  href,
  size = "md",
  fullWidth = false,
  loading = false,
  className = "",
}: SignInWithFacileProps) {
  const { t, i18n } = useTranslation("auth");
  const [pending, setPending] = useState(false);
  const busy = loading || pending;
  const target = withUiLocales(
    href ?? getApiUrl("/auth/oidc"),
    toFacileLocale(i18n.language),
  );
  const classes = `fc-signin${SIZE_CLASS[size]}${fullWidth ? " fc-signin--full" : ""} ${className}`;

  const handleClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
    if (busy) {
      event.preventDefault();
      return;
    }
    const modified =
      event.metaKey ||
      event.ctrlKey ||
      event.shiftKey ||
      event.altKey ||
      event.button !== 0;
    if (!modified) setPending(true);
  };

  return (
    <a
      href={target}
      rel="external"
      aria-busy={busy ? "true" : undefined}
      aria-disabled={busy ? "true" : undefined}
      className={classes.trim()}
      onClick={handleClick}
    >
      {busy ? (
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          aria-hidden="true"
          className="fc-signin__spin"
        >
          <path d="M12 3a9 9 0 1 0 9 9" />
        </svg>
      ) : (
        <svg viewBox="0 0 217 225" fill="currentColor" aria-hidden="true">
          <path d="M21.6 2.10177L23.4 2.70176C19.5 17.4018 27.9 37.8018 42 37.8018C71.4 37.8018 95.1 -0.298234 137.1 0.00176235C170.1 0.00176235 177.3 29.4018 177.3 43.5018V64.8018H131.7L131.4 46.5018C131.4 25.5018 120.3 20.7018 107.7 20.7018C65.7 20.7018 49.8 80.1018 27.3 80.1018C14.7 80.1018 11.4 62.7018 11.4 49.2018C11.4 29.1018 16.5 10.5018 21.6 2.10177ZM53.4 109.202L81.9 46.2018H121.2C109.8 68.1018 100.2 89.4018 91.5 109.202H94.8C129.9 109.202 157.8 117.902 157.8 163.802L112.8 163.202C113.1 121.202 111.9 114.002 95.1 114.002H89.1C60.6 177.902 40.2 225.002 0 225.002L0.3 175.802V109.202H53.4ZM2.7 138.302V175.802C18.3 174.902 32.1 156.902 40.8 137.102L51 114.302C12.3 115.802 2.7 120.602 2.7 138.302Z" />
          <path d="M179.198 223.502V186.902H216.098V223.502H179.198Z" />
        </svg>
      )}
      <span>{t("signIn.label")}</span>
      {busy ? (
        <span className="fc-signin__sr">{t("signIn.loading")}</span>
      ) : null}
    </a>
  );
}
