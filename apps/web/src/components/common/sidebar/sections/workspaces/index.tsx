import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspace";
import useWorkspaceStore from "@/store/workspace";
import { useEffect } from "react";
import AddWorkspace from "./add-workspace";
import WorkspaceItemButton from "./workspace-item-button";

function Workspaces() {
  const { data } = useGetWorkspaces();
  const workspaces = data?.data;
  const { workspace: selectedWorkspace, setWorkspace } = useWorkspaceStore(
    (state) => state,
  );

  useEffect(() => {
    if (data?.data) {
      setWorkspace(data?.data[0]);
    }
  }, [data?.data, setWorkspace]);

  if (!workspaces || !workspaces?.length) {
    // TODO: Add better empty screen
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
            onSelectWorkspace={() => setWorkspace(workspace)}
            isSelected={workspace.id === selectedWorkspace?.id}
          />
        ))}
        <AddWorkspace />
      </div>
    </div>
  );
}

export default Workspaces;
