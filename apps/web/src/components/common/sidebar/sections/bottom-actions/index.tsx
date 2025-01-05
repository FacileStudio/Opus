import { Settings } from "lucide-react";
import SignOutButton from "./sign-out-button";

function BottomActions() {
  return (
    <div className="flex gap-1">
      <button
        type="button"
        className="flex-1 px-2 py-1.5 text-xs text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800 rounded-lg flex items-center justify-center transition-colors"
      >
        <Settings className="w-3 h-3 mr-1" />
        Settings
      </button>
      <SignOutButton />
    </div>
  );
}

export default BottomActions;
