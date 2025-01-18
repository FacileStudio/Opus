import type { Task } from "@/types/workspace";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { format } from "date-fns";
import { Calendar, Flag } from "lucide-react";

interface TaskCardProps {
  task: Task;
}

function TaskCard({ task }: TaskCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: task.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const priorityColors = {
    low: "bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-500",
    medium:
      "bg-yellow-50 text-yellow-700 dark:bg-yellow-500/10 dark:text-yellow-500",
    high: "bg-orange-50 text-orange-700 dark:bg-orange-500/10 dark:text-orange-500",
    urgent: "bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-500",
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="bg-white dark:bg-zinc-800/50 backdrop-blur-sm rounded-lg border border-zinc-200 dark:border-zinc-700/50 p-3 cursor-move hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors shadow-sm"
    >
      <h3 className="font-medium text-zinc-900 dark:text-zinc-100 mb-1">
        {task.title}
      </h3>

      <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-3">
        {task.description}
      </p>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span
            className={`text-xs px-2 py-1 rounded-full ${priorityColors[task.priority as keyof typeof priorityColors]}`}
          >
            <Flag className="w-3 h-3 inline-block mr-1" />
            {task.priority}
          </span>
        </div>

        {task.dueDate && (
          <div className="text-xs text-zinc-600 dark:text-zinc-400 flex items-center">
            <Calendar className="w-3 h-3 mr-1" />
            {format(new Date(task.dueDate), "MMM d")}
          </div>
        )}
      </div>
    </div>
  );
}

export default TaskCard;
