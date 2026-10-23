const ID_KEY = /^(id|.+Id)$/;
const IDS_KEY = /^.+Ids$/;

export function stringifyIds(value: unknown, key?: string): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => stringifyIds(item, key));
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).map(([k, v]) => [k, stringifyIds(v, k)]),
    );
  }
  if (
    typeof value === "number" &&
    key &&
    (ID_KEY.test(key) || IDS_KEY.test(key))
  ) {
    return String(value);
  }
  return value;
}

export function restringify(body: string): string {
  try {
    return JSON.stringify(stringifyIds(JSON.parse(body)));
  } catch {
    return body;
  }
}
