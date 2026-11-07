export const FACILE_LOCALES = ["en", "fr", "es", "de"] as const;

export type FacileLocale = (typeof FACILE_LOCALES)[number];

export const FACILE_LOCALE_NAMES: Record<FacileLocale, string> = {
  en: "English",
  fr: "Français",
  es: "Español",
  de: "Deutsch",
};

export function toFacileLocale(language: string | null | undefined) {
  const base = language?.toLowerCase().split("-")[0];
  return FACILE_LOCALES.find((locale) => locale === base) ?? null;
}

export function withUiLocales(href: string, locale: FacileLocale | null) {
  if (!locale) return href;
  const url = new URL(href, window.location.origin);
  url.searchParams.set("ui_locales", locale);
  return href.startsWith("/") ? `${url.pathname}${url.search}` : url.toString();
}

const REGISTRE = "https://sso.facile.studio";

export function registreSignUpUrl(locale: FacileLocale | null) {
  return locale ? `${REGISTRE}/signup?lang=${locale}` : `${REGISTRE}/signup`;
}

export type LegalDoc = "cgu" | "cgv" | "privacy" | "legal-notice";

export function registreLegalUrl(doc: LegalDoc, locale: FacileLocale | null) {
  const base = `${REGISTRE}/legal/${doc}`;
  return locale ? `${base}?lang=${locale}` : base;
}
