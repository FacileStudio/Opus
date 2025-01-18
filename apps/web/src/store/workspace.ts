import type { Workspace } from "@/types/workspace";
import { create } from "zustand";

const useWorkspaceStore = create<{
  workspace: Workspace | null;
  setWorkspace: (updatedWorkspace: Workspace) => void;
}>((set) => ({
  workspace: null,
  setWorkspace: (updatedWorkspace: Workspace) =>
    set(() => ({ workspace: updatedWorkspace })),
}));

export default useWorkspaceStore;
