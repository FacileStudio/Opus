import { Folder } from "lucide-react";

type WorkspaceItemButtonProps = {
  workspace: {
    id: string;
    name: string;
  };
  isSelected: boolean;
  onSelectWorkspace: (id: string) => void;
};

const buttonStyles: Record<string, string> = {
  true: "bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400",
  false:
    "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800",
};

const iconStyles: Record<string, string> = {
  true: "text-indigo-600 dark:text-indigo-400",
  false: "text-zinc-500 dark:text-zinc-400",
};

function WorkspaceItemButton({
  workspace,
  isSelected,
  onSelectWorkspace,
}: WorkspaceItemButtonProps) {
  return (
    <button
      type="button"
      key={workspace.id}
      onClick={() => onSelectWorkspace(workspace.id)}
      className={`w-full text-left px-3 py-2 rounded-lg flex items-center transition-colors ${buttonStyles[String(isSelected)]}`}
    >
      <Folder className={`w-4 h-4 mr-2 ${iconStyles[String(isSelected)]}`} />
      {workspace.name}
    </button>
  );
}

export default WorkspaceItemButton;
