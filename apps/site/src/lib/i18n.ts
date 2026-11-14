import { type Catalogs, createTranslator, LOCALE_COOKIE, type Locale, resolveLocale } from '@facile/muse';
import { de } from './messages/de';
import { en } from './messages/en';
import { es } from './messages/es';
import { fr } from './messages/fr';

const catalogs: Catalogs<typeof en> = { en, fr, es, de };

let active: () => Locale = () => 'en';

/** The picked locale cookie, then the browser language. */
export function localeSource(): string | null {
	const prefix = `${LOCALE_COOKIE}=`;
	const found = document.cookie.split('; ').find((part) => part.startsWith(prefix));
	return found ? decodeURIComponent(found.slice(prefix.length)) : navigator.language;
}

/** Keeps a locale picked in the footer for the next visit. */
export function rememberLocale(locale: Locale) {
	document.cookie = `${LOCALE_COOKIE}=${locale}; path=/; max-age=31536000; samesite=lax`;
}

/** Points the translator at the locale `provideLocale` resolved. */
export function bindLocale(source: () => Locale) {
	active = () => resolveLocale(source());
}

/** Translates a site message in the current locale. */
export const t = createTranslator(catalogs, () => active());
