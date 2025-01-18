import type { Column } from "@/types/workspace";
import { MoreHorizontal } from "lucide-react";

interface ColumnHeaderProps {
  column: Column;
}

export function ColumnHeader({ column }: ColumnHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <h3 className="font-medium text-zinc-900 dark:text-zinc-100">
          {column.name}
        </h3>
        <span className="text-sm text-zinc-500 dark:text-zinc-500">
          {column.tasks.length}
        </span>
      </div>
      <button
        type="button"
        className="p-1 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-md transition-colors"
      >
        <MoreHorizontal className="w-4 h-4 text-zinc-500 dark:text-zinc-400" />
      </button>
    </div>
  );
}
