import AddWorkspace from "./add-workspace";
import WorkspaceItemButton from "./workspace-item-button";

const workspaces = [
  {
    id: "1",
    name: "Personal Workspace",
  },
];

function Workspaces() {
  const selectedWorkspace = {
    id: "1",
    name: "Personal Workspace",
  };

  return (
    <div>
      <h2 className="text-xs font-semibold text-zinc-500 dark:text-zinc-400 uppercase tracking-wider mb-2 px-3">
        Workspaces
      </h2>
      <div className="space-y-1">
        {workspaces.map((workspace) => (
          <WorkspaceItemButton
            key={workspace.id}
            workspace={workspace}
            onSelectWorkspace={() => console.log("Selected")}
            isSelected={workspace.id === selectedWorkspace.id}
          />
        ))}

        <AddWorkspace />
      </div>
    </div>
  );
}

export default Workspaces;
