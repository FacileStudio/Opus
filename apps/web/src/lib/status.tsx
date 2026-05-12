import {
  CheckCircle as CheckCircle2,
  ClockCircle as Clock,
} from "@solar-icons/react";

export function getStatusIcon(status: "active" | "pending") {
  switch (status) {
    case "active":
      return <CheckCircle2 size={16} className="text-success-foreground" />;
    case "pending":
      return <Clock size={16} className="text-warning-foreground" />;
  }
}

export function getStatusText(status: "active" | "pending") {
  switch (status) {
    case "active":
      return "Active";
    case "pending":
      return "Pending";
  }
}
