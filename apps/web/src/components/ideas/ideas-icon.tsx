import { Icon } from "@iconify/react";
import { IDEAS_ICON } from "@/lib/ideas";

export function IdeasIcon({
  className,
  size,
}: {
  className?: string;
  size?: number;
}) {
  return (
    <Icon
      icon={IDEAS_ICON}
      className={className}
      width={size}
      height={size}
      aria-hidden="true"
    />
  );
}
