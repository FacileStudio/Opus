import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspace";
import useWorkspaceStore from "@/store/workspace";
import { FilePlus } from "lucide-react";
import type React from "react";
import { useEffect, useState } from "react";
import AddWorkspace from "./components/add-workspace";
import WorkspaceItem from "./components/workspace-item";

function Workspaces() {
  const { data } = useGetWorkspaces();
  const workspaces = data?.data;
  const { workspace: selectedWorkspace, setWorkspace } = useWorkspaceStore(
    (state) => state,
  );

  const [expandedWorkspaces, setExpandedWorkspaces] = useState<string[]>([
    selectedWorkspace?.id || "",
  ]);

  const handleExpandWorkspace = (workspaceId: string) => {
    setExpandedWorkspaces((prev) =>
      prev.includes(workspaceId)
        ? prev.filter((id) => id !== workspaceId)
        : [...prev, workspaceId],
    );
  };

  const handleSelectWorkspace = (
    event:
      | React.MouseEvent<HTMLDivElement>
      | React.KeyboardEvent<HTMLDivElement>,
    workspaceId: string,
  ) => {
    event.stopPropagation();

    const newWorkspace = workspaces?.find(
      (workspace) => workspace.id === workspaceId,
    );

    if (!newWorkspace) return;

    setWorkspace(newWorkspace);

    if (!expandedWorkspaces.includes(workspaceId)) {
      setExpandedWorkspaces((prev) => [...prev, workspaceId]);
    }
  };

  useEffect(() => {
    if (data?.data && !selectedWorkspace) {
      setWorkspace(data?.data[0]);
    }
  }, [data?.data, setWorkspace, selectedWorkspace]);

  if (!workspaces || !workspaces.length) {
    return (
      <>
        <div className="flex items-center justify-center px-3 pt-8 text-center">
          <div className="text-sm text-zinc-500 dark:text-zinc-400">
            <FilePlus className="mx-auto mb-5 text-zinc-400 dark:text-zinc-500 w-8 h-8" />
            <p className="text-base text-zinc-700 dark:text-zinc-200 font-medium">
              No workspaces yet
            </p>
            <p className="text-sm text-zinc-400 dark:text-zinc-500 mt-2">
              Create a workspace to organize your projects.
            </p>
          </div>
        </div>
        <AddWorkspace />
      </>
    );
  }

  return (
    <div>
      <h2 className="text-xs font-semibold text-zinc-500 dark:text-zinc-400 uppercase tracking-wider mb-2 px-3">
        Workspaces
      </h2>
      <div className="space-y-1">
        {workspaces.length > 0 &&
          workspaces.map((workspace) => (
            <div key={workspace.id} className="space-y-1">
              <WorkspaceItem
                workspace={workspace}
                isExpanded={expandedWorkspaces.includes(workspace.id)}
                onSelectWorkspace={handleSelectWorkspace}
                onExpandWorkspace={handleExpandWorkspace}
              />
            </div>
          ))}
        <AddWorkspace />
      </div>
    </div>
  );
}

export default Workspaces;
