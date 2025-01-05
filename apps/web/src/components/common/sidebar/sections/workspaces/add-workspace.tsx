import { Plus } from "lucide-react";

function AddWorkspace() {
  return (
    <button
      type="button"
      className="w-full text-left px-3 py-2 text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800 rounded-lg flex items-center transition-colors"
    >
      <Plus className="w-4 h-4 mr-2 text-zinc-500 dark:text-zinc-400" />
      Add Workspace
    </button>
  );
}

export default AddWorkspace;
