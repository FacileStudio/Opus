export const USER_COLORS = [
  "#AD9EF0",
  "#F09ED6",
  "#EE7E89",
  "#EEB47E",
  "#A9EE7E",
  "#7EEEDB",
] as const;

export type UserColor = (typeof USER_COLORS)[number];

export const USER_COLOR_LABELS: Record<UserColor, string> = {
  "#AD9EF0": "Purple",
  "#F09ED6": "Pink",
  "#EE7E89": "Red",
  "#EEB47E": "Orange",
  "#A9EE7E": "Green",
  "#7EEEDB": "Blue",
};

export function normalizeUserColor(
  color: string | null | undefined,
): UserColor {
  if (!color) return USER_COLORS[0];
  const upper = color.trim().toUpperCase();
  const withHash = upper.startsWith("#") ? upper : `#${upper}`;
  return (USER_COLORS as readonly string[]).includes(withHash)
    ? (withHash as UserColor)
    : USER_COLORS[0];
}
