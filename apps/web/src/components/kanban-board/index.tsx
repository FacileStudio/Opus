import generateProject from "@/lib/workspace/generate-project";
import useProjectStore from "@/store/project";
import {
  DndContext,
  type DragEndEvent,
  DragOverlay,
  type DragStartEvent,
  type UniqueIdentifier,
  closestCorners,
} from "@dnd-kit/core";
import { useState } from "react";
import Column from "./column";
import TaskCard from "./task-card";

function KanbanBoard() {
  const [project, setProject] = useState(
    generateProject({
      projectId: "sample-1",
      workspaceId: "workspace-1",
      tasksPerColumn: 4,
    }),
  );
  const { project: selectedProject } = useProjectStore();
  const [activeId, setActiveId] = useState<UniqueIdentifier | null>(null);

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id);
  };

  // TODO: Simplify this function
  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over) return;

    const activeId = active.id.toString();
    const overId = over.id.toString();

    const sourceColumn = project.columns.find((col) =>
      col.tasks.some((task) => task.id === activeId),
    );

    const destinationColumn = project.columns.find((col) => {
      if (col.id === overId) return true;
      return col.tasks.some((task) => task.id === overId);
    });

    if (!sourceColumn || !destinationColumn) return;

    setProject((currentProject) => {
      const updatedColumns = [...currentProject.columns];

      const sourceColumnIndex = updatedColumns.findIndex(
        (col) => col.id === sourceColumn.id,
      );
      const destinationColumnIndex = updatedColumns.findIndex(
        (col) => col.id === destinationColumn.id,
      );

      const sourceTaskIndex = sourceColumn.tasks.findIndex(
        (task) => task.id === activeId,
      );
      const task = sourceColumn.tasks[sourceTaskIndex];

      updatedColumns[sourceColumnIndex] = {
        ...sourceColumn,
        tasks: sourceColumn.tasks.filter((t) => t.id !== activeId),
      };

      if (sourceColumn.id === destinationColumn.id) {
        const destinationIndex = destinationColumn.tasks.findIndex(
          (task) => task.id === overId,
        );
        const newTasks = [...destinationColumn.tasks];
        newTasks.splice(sourceTaskIndex, 1);
        newTasks.splice(destinationIndex, 0, task);

        updatedColumns[destinationColumnIndex] = {
          ...destinationColumn,
          tasks: newTasks,
        };
      } else {
        const updatedTask = { ...task, status: destinationColumn.id };

        if (overId === destinationColumn.id) {
          updatedColumns[destinationColumnIndex] = {
            ...destinationColumn,
            tasks: [...destinationColumn.tasks, updatedTask],
          };
        } else {
          const destinationIndex = destinationColumn.tasks.findIndex(
            (task) => task.id === overId,
          );
          const newTasks = [...destinationColumn.tasks];
          newTasks.splice(destinationIndex, 0, updatedTask);

          updatedColumns[destinationColumnIndex] = {
            ...destinationColumn,
            tasks: newTasks,
          };
        }
      }

      return {
        ...currentProject,
        columns: updatedColumns,
      };
    });

    setActiveId(null);
  };

  const activeTask = activeId
    ? project.columns
        .flatMap((col) => col.tasks)
        .find((task) => task.id === activeId)
    : null;

  return (
    <DndContext
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="h-full flex flex-col">
        <header className="mb-6 space-y-6 shrink-0 px-4 md:px-0">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
              {selectedProject?.name}
            </h1>
          </div>
        </header>

        <div className="flex-1 relative min-h-0">
          <div className="flex gap-6 overflow-x-auto pb-6 px-4 md:px-6 h-full snap-x snap-mandatory scrollbar-thin scrollbar-track-zinc-100 scrollbar-thumb-zinc-300 dark:scrollbar-track-zinc-900 dark:scrollbar-thumb-zinc-700">
            {project.columns.map((column) => (
              <Column key={column.id} column={column} />
            ))}
          </div>
        </div>
      </div>

      <DragOverlay>
        {activeTask ? (
          <div className="transform rotate-3">
            <TaskCard task={activeTask} />
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}

export default KanbanBoard;
