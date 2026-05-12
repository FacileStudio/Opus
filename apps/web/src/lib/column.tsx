import {
  Archive,
  CheckCircle as CheckCircle2,
  Record as Circle,
  Stopwatch as CircleDashed,
  RecordCircle as CircleDot,
  Magnifier as Search,
} from "@solar-icons/react";

export const getColumnIcon = (columnId: string, isFinal?: boolean) => {
  switch (columnId) {
    case "to-do":
      return <Circle size={16} className="text-muted-foreground" />;
    case "in-progress":
      return <CircleDot size={16} className="text-muted-foreground" />;
    case "in-review":
      return <Search size={16} className="text-muted-foreground" />;
    case "done":
      return <CheckCircle2 size={16} className="text-muted-foreground" />;
    case "archived":
      return <Archive size={16} className="text-muted-foreground" />;
    case "planned":
      return <CircleDashed size={16} className="text-muted-foreground" />;
    default:
      return isFinal ? (
        <CheckCircle2 size={16} className="text-muted-foreground" />
      ) : (
        <Circle size={16} className="text-muted-foreground" />
      );
  }
};
