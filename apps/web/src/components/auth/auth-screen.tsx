import { useTranslation } from "react-i18next";
import { Logo } from "@/components/common/logo";
import { i18n as appI18n, LOCALE_STORAGE_KEY, resolveLocale } from "@/lib/i18n";
import {
  FACILE_LOCALE_NAMES,
  FACILE_LOCALES,
  type LegalDoc,
  registreLegalUrl,
  registreSignUpUrl,
  toFacileLocale,
} from "./facile-locale";
import { SignInWithFacile } from "./sign-in-with-facile";

type AuthScreenProps = {
  notice?: React.ReactNode;
  noticeTone?: "info" | "error";
};

const LEGAL_LINKS: Array<{ doc: LegalDoc; key: string }> = [
  { doc: "cgu", key: "cgu" },
  { doc: "cgv", key: "cgv" },
  { doc: "privacy", key: "privacy" },
  { doc: "legal-notice", key: "legalNotice" },
];

export function AuthScreen({ notice, noticeTone = "info" }: AuthScreenProps) {
  const { t, i18n } = useTranslation("auth");
  const locale = toFacileLocale(i18n.language);

  const onLocaleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const next = resolveLocale(event.target.value);
    void appI18n.changeLanguage(next);
    document.documentElement.lang = next;
    try {
      localStorage.setItem(LOCALE_STORAGE_KEY, next);
    } catch {
      return;
    }
  };

  return (
    <div className="fc-auth">
      <div aria-hidden="true" />
      <section className="fc-auth__card" aria-labelledby="fc-auth-title">
        <div className="fc-auth__head">
          <Logo to="/" />
          <h1 id="fc-auth-title" className="fc-auth__title">
            Opus
          </h1>
          <p className="fc-auth__tagline">{t("signIn.tagline")}</p>
        </div>
        {notice ? (
          <div
            role={noticeTone === "error" ? "alert" : "status"}
            className={`fc-auth__notice${noticeTone === "error" ? " fc-auth__notice--error" : ""}`}
          >
            {notice}
          </div>
        ) : null}
        <SignInWithFacile size="lg" fullWidth />
        <p className="fc-auth__more">
          {t("signIn.newHere")}{" "}
          <a
            className="fc-auth__link"
            href={registreSignUpUrl(locale)}
            rel="external"
          >
            {t("signIn.createAccount")}
          </a>
        </p>
      </section>
      <footer className="fc-auth__foot">
        <nav aria-label={t("legal.nav")}>
          <ul className="fc-auth__legal">
            {LEGAL_LINKS.map(({ doc, key }) => (
              <li key={doc}>
                <a href={registreLegalUrl(doc, locale)} rel="external">
                  {t(`legal.${key}`)}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <select
          className="fc-auth__locale"
          aria-label={t("signIn.language")}
          value={locale ?? "en"}
          onChange={onLocaleChange}
        >
          {FACILE_LOCALES.map((code) => (
            <option key={code} value={code}>
              {FACILE_LOCALE_NAMES[code]}
            </option>
          ))}
        </select>
      </footer>
    </div>
  );
}
