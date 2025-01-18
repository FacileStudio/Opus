import type { Column as ColumnType } from "@/types/workspace";
import { ColumnDropzone } from "./column-dropzone";
import { ColumnHeader } from "./column-header";

interface ColumnProps {
  column: ColumnType;
}

function Column({ column }: ColumnProps) {
  return (
    <div className="flex flex-col w-[calc(100vw-2rem)] md:w-80 shrink-0 bg-zinc-50/50 dark:bg-zinc-900/50 backdrop-blur-sm rounded-lg border border-zinc-200 dark:border-zinc-800/50 snap-center h-full shadow-sm">
      <div className="p-3 border-b border-zinc-200 dark:border-zinc-800/50">
        <ColumnHeader column={column} />
      </div>
      <div className="p-3 overflow-y-auto flex-1">
        <ColumnDropzone column={column} />
      </div>
    </div>
  );
}

export default Column;
