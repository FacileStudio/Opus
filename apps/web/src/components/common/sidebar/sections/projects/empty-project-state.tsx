import { Button } from "@/components/ui/button";
import useGetWorkspaces from "@/hooks/queries/workspace/use-get-workspace";
import { motion } from "framer-motion";
import { Layout, Plus } from "lucide-react";
import { useState } from "react";
import { CreateWorkspaceModal } from "../workspaces/components/create-workspace-modal";
import CreateProjectModal from "./create-project-modal";

export function BoardEmptyState() {
  const { data: workspaces } = useGetWorkspaces();
  const [isCreateWorkspaceOpen, setIsCreateWorkspaceOpen] = useState(false);
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false);

  return (
    <div className="h-full flex flex-col items-center justify-center p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-md w-full text-center"
      >
        <div className="w-20 h-20 mx-auto mb-8 relative">
          <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/20 to-purple-500/20 dark:from-indigo-500/10 dark:to-purple-500/10 rounded-2xl transform rotate-6" />
          <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/20 to-purple-500/20 dark:from-indigo-500/10 dark:to-purple-500/10 rounded-2xl transform -rotate-6" />
          <div className="relative w-full h-full bg-gradient-to-br from-indigo-500 to-purple-500 rounded-2xl flex items-center justify-center shadow-xl">
            <Layout className="w-10 h-10 text-white" />
          </div>
        </div>

        <h2 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100 mb-3">
          {workspaces?.data && workspaces?.data.length === 0
            ? "No Workspace Selected"
            : "No Project Selected"}
        </h2>

        <p className="text-zinc-600 dark:text-zinc-400 mb-8 max-w-sm mx-auto">
          {workspaces?.data && workspaces?.data.length === 0
            ? "Get started by creating your first workspace and organizing your projects."
            : "Select a project from the sidebar or create a new one to get started."}
        </p>

        <div className="space-y-3">
          {workspaces?.data && workspaces?.data.length === 0 ? (
            <Button
              onClick={() => setIsCreateWorkspaceOpen(true)}
              className="bg-indigo-600 text-white hover:bg-indigo-500 dark:bg-indigo-500 dark:hover:bg-indigo-400 w-full"
            >
              <Plus className="w-4 h-4 mr-2" />
              Create First Workspace
            </Button>
          ) : (
            <>
              <Button
                onClick={() => setIsCreateProjectOpen(true)}
                className="bg-indigo-600 text-white hover:bg-indigo-500 dark:bg-indigo-500 dark:hover:bg-indigo-400 w-full"
              >
                <Plus className="w-4 h-4 mr-2" />
                Create New Project
              </Button>
            </>
          )}
        </div>

        <div className="mt-12 pt-8 border-t border-zinc-200 dark:border-zinc-800">
          <h3 className="text-sm font-medium text-zinc-900 dark:text-zinc-100 mb-4">
            Quick Tips
          </h3>
          <div className="grid gap-4 text-sm">
            <div className="p-3 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-left">
              <p className="text-zinc-600 dark:text-zinc-400">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  Create workspaces
                </span>{" "}
                to organize different areas of work
              </p>
            </div>
            <div className="p-3 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-left">
              <p className="text-zinc-600 dark:text-zinc-400">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  Add projects
                </span>{" "}
                to track specific initiatives or goals
              </p>
            </div>
            <div className="p-3 rounded-lg bg-zinc-50 dark:bg-zinc-900 text-left">
              <p className="text-zinc-600 dark:text-zinc-400">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  Invite team members
                </span>{" "}
                to collaborate on tasks together
              </p>
            </div>
          </div>
        </div>
      </motion.div>
      <CreateWorkspaceModal
        open={isCreateWorkspaceOpen}
        onClose={() => setIsCreateWorkspaceOpen(false)}
      />
      <CreateProjectModal
        open={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
      />
    </div>
  );
}
