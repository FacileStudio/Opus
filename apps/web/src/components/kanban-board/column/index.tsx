import { useState } from "react";
import type { ProjectWithTasks } from "@/types/project";
import { ColumnDropzone } from "./column-dropzone";
import { ColumnHeader } from "./column-header";

type ColumnProps = {
  column: ProjectWithTasks["columns"][number];
};

function Column({ column }: ColumnProps) {
  const [isDropzoneOver, setIsDropzoneOver] = useState(false);

  return (
    <div
      className={`group relative flex h-full min-h-0 w-full flex-col rounded-xl transition-all duration-300 ease-out ${
        isDropzoneOver ? "bg-accent ring-2 ring-ring/30" : "bg-muted"
      }`}
    >
      <div className="shrink-0 px-3 py-2">
        <ColumnHeader column={column} />
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden px-2 py-1 [-webkit-overflow-scrolling:touch]">
        <ColumnDropzone column={column} onIsOverChange={setIsDropzoneOver} />
      </div>
    </div>
  );
}

export default Column;
