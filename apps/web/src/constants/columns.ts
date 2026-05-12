import {
  CheckCircle as Check,
  Record as Circle,
  RecordCircle as CircleDot,
  Magnifier as Search,
} from "@solar-icons/react";

export const DEFAULT_COLUMNS = [
  { id: "to-do", name: "To Do", icon: Circle },
  { id: "in-progress", name: "In Progress", icon: CircleDot },
  { id: "in-review", name: "In Review", icon: Search },
  { id: "done", name: "Done", icon: Check },
] as const;
