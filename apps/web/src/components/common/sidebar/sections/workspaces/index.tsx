import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspace";
import AddWorkspace from "./add-workspace";
import WorkspaceItemButton from "./workspace-item-button";

const selectedWorkspace = {
  id: "1",
};

function Workspaces() {
  const { data } = useGetWorkspaces();
  const workspaces = data?.data;

  if (!workspaces || !workspaces?.length) {
    return <div>You don't have any workspaces</div>;
  }

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
