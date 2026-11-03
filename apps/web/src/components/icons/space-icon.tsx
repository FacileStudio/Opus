import { type IconProps, ShareCircle } from "@solar-icons/react";

/**
 * The suite-wide glyph for a space: Solar share-circle, bold duotone.
 */
export function SpaceIcon(props: Omit<IconProps, "ref">) {
  return <ShareCircle weight="BoldDuotone" {...props} />;
}
