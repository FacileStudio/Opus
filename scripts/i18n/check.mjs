import {
  defaultLocale,
  flattenLocale,
  formatKeyList,
  getValueAtKey,
  loadLocales,
  setValueAtKey,
  writeJson,
} from "./shared.mjs";

const args = process.argv.slice(2);
const shouldFix = args.includes("--fix");
const localeFilter = args.find((arg) => arg !== "--fix");

const { locales, reference } = await loadLocales();
const referenceKeys = flattenLocale(reference.data);
const targetLocales = locales.filter(({ locale }) => locale !== defaultLocale);
const filteredLocales = localeFilter
  ? targetLocales.filter(({ locale }) => locale === localeFilter)
  : targetLocales;

if (localeFilter && filteredLocales.length === 0) {
  console.error(`No locale found for "${localeFilter}".`);
  process.exit(1);
}

const pluralSuffix = /_(zero|one|two|few|many|other)$/u;

function pluralBase(key) {
  return key.replace(pluralSuffix, "");
}

function requiredForms(locale) {
  const categories = new Intl.PluralRules(locale).resolvedOptions()
    .pluralCategories;

  return categories.filter(
    (category) => category !== "many" || categories.includes("few"),
  );
}

function groupByBase(keys) {
  const groups = new Map();

  for (const key of keys) {
    const base = pluralBase(key);
    groups.set(base, [...(groups.get(base) ?? []), key]);
  }

  return groups;
}

const referenceGroups = groupByBase(referenceKeys);

let hasIssues = false;

for (const locale of filteredLocales) {
  const localeKeys = flattenLocale(locale.data);
  const localeGroups = groupByBase(localeKeys);
  const missing = new Set();
  const extra = new Set();

  for (const [base, keys] of referenceGroups) {
    const localKeys = localeGroups.get(base);
    const isPlural = keys.some((key) => pluralSuffix.test(key));

    if (!localKeys) {
      for (const key of keys) {
        missing.add(key);
      }
      continue;
    }

    if (isPlural && localKeys.some((key) => pluralSuffix.test(key))) {
      for (const form of requiredForms(locale.locale)) {
        const present =
          localKeys.includes(`${base}_${form}`) ||
          (form === "one" && localKeys.includes(base));

        if (!present) {
          missing.add(`${base}_${form}`);
        }
      }
    }
  }

  for (const [base, keys] of localeGroups) {
    if (!referenceGroups.has(base)) {
      for (const key of keys) {
        extra.add(key);
      }
    }
  }

  if (missing.size === 0 && extra.size === 0) {
    console.log(`${locale.locale}: OK`);
    continue;
  }

  hasIssues = true;
  console.log(`${locale.locale}:`);

  if (missing.size > 0) {
    console.log("  Missing keys:");
    for (const key of formatKeyList(missing)) {
      console.log(`    - ${key}`);
      if (shouldFix) {
        setValueAtKey(
          locale.data,
          key,
          getValueAtKey(reference.data, key) ??
            getValueAtKey(reference.data, `${pluralBase(key)}_other`),
        );
      }
    }
  }

  if (extra.size > 0) {
    console.log("  Extra keys:");
    for (const key of formatKeyList(extra)) {
      console.log(`    - ${key}`);
    }
  }

  if (shouldFix && missing.size > 0) {
    await writeJson(locale.path, locale.data);
    console.log("  Added missing keys from en-US.");
  }
}

if (!hasIssues) {
  console.log("All locale files are in sync with en-US.");
  process.exit(0);
}

process.exit(shouldFix ? 0 : 1);
