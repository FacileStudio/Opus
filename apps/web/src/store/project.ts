import type { Project } from "@/types/project";
import { create } from "zustand";

const useProjectStore = create<{
  project: Project | null;
  setProject: (updatedProject: Project) => void;
}>((set) => ({
  project: null,
  setProject: (updatedProject) => set(() => ({ project: updatedProject })),
}));

export default useProjectStore;
