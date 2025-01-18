import { motion } from "framer-motion";
import { SidebarContent } from "./sidebar-content";
import SidebarFooter from "./sidebar-footer";
import { SidebarHeader } from "./sidebar-header";

export function Sidebar() {
  return (
    <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
      <div className="flex w-64 bg-white dark:bg-zinc-900 border-r border-zinc-200 dark:border-zinc-800 h-screen flex-col">
        <SidebarHeader />
        <SidebarContent />
        <SidebarFooter />
      </div>
    </motion.div>
  );
}
