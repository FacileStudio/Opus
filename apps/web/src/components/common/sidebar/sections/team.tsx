import { Users } from "lucide-react";

function Team() {
  return (
    <div>
      <h2 className="text-xs font-semibold text-zinc-500 dark:text-zinc-400 uppercase tracking-wider mb-2 px-3">
        Team
      </h2>
      <button
        type="button"
        className="w-full text-left px-3 py-2 text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800 rounded-lg flex items-center transition-colors"
      >
        <Users className="w-4 h-4 mr-2 text-zinc-500 dark:text-zinc-400" />
        Manage Team
      </button>
    </div>
  );
}

export default Team;
