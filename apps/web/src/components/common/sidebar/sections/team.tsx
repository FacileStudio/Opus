import { Users } from "lucide-react";

function Team() {
  return (
    <div>
      <h2 className="text-xs font-semibold text-zinc-500 dark:text-zinc-400 uppercase tracking-wider mb-2 px-3">
        Team
      </h2>
      <button
        type="button"
        className="w-full text-left px-2 py-1.5 text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800/50 rounded-md flex items-center transition-all group"
      >
        <Users className="w-4 h-4 mr-2 text-zinc-400 dark:text-zinc-500 group-hover:text-zinc-500 dark:group-hover:text-zinc-400" />
        <span className="text-sm font-medium group-hover:text-zinc-900 dark:group-hover:text-zinc-300">
          Manage Team
        </span>
      </button>
    </div>
  );
}

export default Team;
