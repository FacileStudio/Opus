import { create } from "zustand";
import { persist } from "zustand/middleware";

interface ThemeStore {
  theme: "light" | "dark" | "system";
  setTheme: (theme: "light" | "dark" | "system") => void;
}

const useTheme = create<ThemeStore>()(
  persist(
    (set) => ({
      theme: "dark",
      setTheme: (theme: "light" | "dark" | "system") => set({ theme }),
    }),
    { name: "theme-storage" },
  ),
);

export default useTheme;
