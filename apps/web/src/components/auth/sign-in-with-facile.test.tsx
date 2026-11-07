import { cleanup, render, screen } from "@testing-library/react";
import { I18nextProvider } from "react-i18next";
import { afterEach, describe, expect, it } from "vitest";
import { i18n } from "@/lib/i18n";
import { toFacileLocale, withUiLocales } from "./facile-locale";
import { SignInWithFacile } from "./sign-in-with-facile";

const LABELS = {
  "en-US": "Sign in with Facile",
  "fr-FR": "Se connecter avec Facile",
  "es-ES": "Iniciar sesión con Facile",
  "de-DE": "Mit Facile anmelden",
} as const;

describe("SignInWithFacile", () => {
  afterEach(cleanup);

  it.each(
    Object.entries(LABELS),
  )("says the canonical label in %s", async (locale, label) => {
    await i18n.changeLanguage(locale);
    render(
      <I18nextProvider i18n={i18n}>
        <SignInWithFacile />
      </I18nextProvider>,
    );

    const link = screen.getByRole("link", { name: label });
    expect(link.getAttribute("href")).toContain(
      `ui_locales=${locale.slice(0, 2)}`,
    );
    expect(link.getAttribute("href")).toContain("/auth/oidc");
  });
});

describe("facile locale", () => {
  it("keeps only the four suite locales", () => {
    expect(toFacileLocale("fr-FR")).toBe("fr");
    expect(toFacileLocale("de")).toBe("de");
    expect(toFacileLocale("nl-NL")).toBeNull();
    expect(toFacileLocale(undefined)).toBeNull();
  });

  it("adds ui_locales only for a supported locale", () => {
    expect(withUiLocales("/api/auth/oidc", "es")).toBe(
      "/api/auth/oidc?ui_locales=es",
    );
    expect(withUiLocales("/api/auth/oidc", null)).toBe("/api/auth/oidc");
  });
});
