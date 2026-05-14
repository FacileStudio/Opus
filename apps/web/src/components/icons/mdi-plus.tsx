import { mdiPlus } from "@mdi/js";
import Icon from "@mdi/react";

type MdiPlusProps = {
  size?: number;
  className?: string;
  weight?: string;
};

export function MdiPlus({ size, className }: MdiPlusProps) {
  return (
    <Icon
      path={mdiPlus}
      size={size ? `${size}px` : undefined}
      className={className}
    />
  );
}
