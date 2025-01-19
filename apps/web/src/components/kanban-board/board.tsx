import Column from "./column";

// TODO: Remove this component
function Board({ project, selectedProject }) {
  return (
    <div className="h-full flex flex-col">
      <header className="mb-6 space-y-6 flex-shrink-0 px-4 md:px-0">
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
  );
}
export default Board;
