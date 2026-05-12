import {
  AltArrowDown as ChevronDown,
  DoubleAltArrowUp as ChevronsUp,
  AltArrowUp as ChevronUp,
  DangerCircle as CircleAlert,
  MinusCircle as Minus,
} from "@solar-icons/react";

export function getPriorityIcon(priority: string) {
  switch (priority) {
    case "urgent":
      return <CircleAlert size={12} className="text-destructive-foreground" />;
    case "high":
      return <ChevronsUp size={12} className="text-warning-foreground" />;
    case "medium":
      return <ChevronUp size={12} className="text-warning-foreground/80" />;
    case "low":
      return <ChevronDown size={12} className="text-info-foreground/85" />;
    case "no-priority":
      return <Minus size={12} className="text-muted-foreground" />;
    default:
      return <Minus size={12} className="text-muted-foreground" />;
  }
}
