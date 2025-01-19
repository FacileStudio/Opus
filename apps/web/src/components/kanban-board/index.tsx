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
import { useEffect, useState } from "react";
import Board from "./board";
import TaskCard from "./task-card";

function KanbanBoard() {
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null);
  const [project, setProject] = useState(
    generateProject({
      projectId: "sample-1",
      workspaceId: "workspace-1",
      tasksPerColumn: 4,
    }),
  );
  const { project: selectedProject } = useProjectStore();
  const [activeId, setActiveId] = useState<UniqueIdentifier | null>(null);

  useEffect(() => {
    if (selectedProject) {
      // TODO: Clean this up by creating a hook for websockets and using Zustand for projects
      const freshWebsocket = new WebSocket(
        `http://localhost:1337/task/${selectedProject?.id}`,
      );
      setWebSocket(freshWebsocket);

      freshWebsocket.onmessage = (event) => {
        console.log("OPENED!", event.data);

        setProject(JSON.parse(event.data));
      };
    }
  }, [selectedProject?.id]);

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

        webSocket?.send(
          JSON.stringify({
            type: "UPDATE_TASK",
            ...updatedTask,
          }),
        );

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
      <Board project={project} selectedProject={selectedProject} />
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
