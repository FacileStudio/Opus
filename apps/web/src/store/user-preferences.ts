import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

type UserPreferencesStore = {
  theme: "light" | "dark" | "system";
  setTheme: (
    theme: "light" | "dark" | "system",
    coordinates?: { x: number; y: number },
  ) => void;

  viewMode: "board" | "list";
  setViewMode: (mode: "board" | "list") => void;

  compactMode: boolean;
  setCompactMode: (compact: boolean) => void;

  showTaskNumbers: boolean;
  setShowTaskNumbers: (show: boolean) => void;
  toggleTaskNumbers: () => void;
  showActors: boolean;
  setShowActors: (show: boolean) => void;
  toggleActors: () => void;
  showDueDates: boolean;
  setShowDueDates: (show: boolean) => void;
  toggleDueDates: () => void;
  showLabels: boolean;
  setShowLabels: (show: boolean) => void;
  toggleLabels: () => void;
  showPriority: boolean;
  setShowPriority: (show: boolean) => void;
  togglePriority: () => void;
  resetDisplayPreferences: () => void;

  projectListView: "card" | "list";
  setProjectListView: (view: "card" | "list") => void;

  myTasksView: "list" | "columns";
  setMyTasksView: (view: "list" | "columns") => void;

  sidebarDefaultOpen: boolean;
  setSidebarDefaultOpen: (open: boolean) => void;

  weekStartsOn: 0 | 1;
  setWeekStartsOn: (weekStartsOn: 0 | 1) => void;
};

export const useUserPreferencesStore = create<UserPreferencesStore>()(
  persist(
    (set) => ({
      theme: "dark",
      setTheme: (
        theme: "light" | "dark" | "system",
        coordinates?: { x: number; y: number },
      ) => {
        if (coordinates) {
          document.documentElement.style.setProperty(
            "--x",
            `${coordinates.x}%`,
          );
          document.documentElement.style.setProperty(
            "--y",
            `${coordinates.y}%`,
          );
        } else {
          document.documentElement.style.removeProperty("--x");
          document.documentElement.style.removeProperty("--y");
        }

        if ("startViewTransition" in document) {
          document.startViewTransition(() => {
            set({ theme });
          });
        } else {
          set({ theme });
        }
      },

      viewMode: "board",
      setViewMode: (mode) => set({ viewMode: mode }),

      compactMode: false,
      setCompactMode: (compact) => set({ compactMode: compact }),

      showTaskNumbers: true,
      setShowTaskNumbers: (show) => set({ showTaskNumbers: show }),
      toggleTaskNumbers: () =>
        set((state) => ({ showTaskNumbers: !state.showTaskNumbers })),
      showActors: true,
      setShowActors: (show) => set({ showActors: show }),
      toggleActors: () => set((state) => ({ showActors: !state.showActors })),
      showDueDates: true,
      setShowDueDates: (show) => set({ showDueDates: show }),
      toggleDueDates: () =>
        set((state) => ({ showDueDates: !state.showDueDates })),
      showLabels: true,
      setShowLabels: (show) => set({ showLabels: show }),
      toggleLabels: () => set((state) => ({ showLabels: !state.showLabels })),
      showPriority: true,
      setShowPriority: (show) => set({ showPriority: show }),
      togglePriority: () =>
        set((state) => ({ showPriority: !state.showPriority })),
      resetDisplayPreferences: () =>
        set({
          showActors: true,
          showDueDates: true,
          showLabels: true,
          showTaskNumbers: true,
          showPriority: true,
        }),

      projectListView: "card",
      setProjectListView: (view) => set({ projectListView: view }),

      myTasksView: "list",
      setMyTasksView: (view) => set({ myTasksView: view }),

      sidebarDefaultOpen: true,
      setSidebarDefaultOpen: (open) => set({ sidebarDefaultOpen: open }),

      weekStartsOn: 0,
      setWeekStartsOn: (weekStartsOn) => set({ weekStartsOn }),
    }),
    {
      name: "user-preferences",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
