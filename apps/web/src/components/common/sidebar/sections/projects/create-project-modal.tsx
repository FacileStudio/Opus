import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import icons from "@/constants/project-icons";
import useCreateProject from "@/hooks/mutations/project/use-create-project";
import { cn } from "@/lib/cn";
import useWorkspaceStore from "@/store/workspace";
import * as Dialog from "@radix-ui/react-dialog";
import { useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useState } from "react";

type CreateProjectModalProps = {
  open: boolean;
  onClose: () => void;
};

function CreateProjectModal({ open, onClose }: CreateProjectModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [selectedIcon, setSelectedIcon] = useState("Layout");
  const queryClient = useQueryClient();
  const { workspace } = useWorkspaceStore();
  const { mutateAsync } = useCreateProject({
    name,
    description,
    workspaceId: workspace?.id ?? "",
    icon: selectedIcon,
  });
  const IconComponent = icons[selectedIcon as keyof typeof icons];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    await mutateAsync();
    await queryClient.invalidateQueries({ queryKey: ["projects"] });

    setName("");
    setDescription("");
    setSelectedIcon("Layout");
    onClose();
  };

  return (
    <Dialog.Root open={open} onOpenChange={onClose}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/40 backdrop-blur-xs" />
        <Dialog.Content className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-md">
          <div className="bg-white dark:bg-zinc-900 rounded-lg shadow-xl">
            <div className="flex items-center justify-between p-4 border-b border-zinc-200 dark:border-zinc-800">
              <Dialog.Title className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
                New Project
              </Dialog.Title>
              <Dialog.Close className="text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-300">
                <X size={20} />
              </Dialog.Close>
            </div>

            <form onSubmit={handleSubmit} className="p-4">
              <div className="mb-4">
                <label
                  htmlFor="projectName"
                  aria-label="Project name"
                  className="block text-sm font-medium text-zinc-900 dark:text-zinc-300 mb-1"
                >
                  Project Name
                </label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Designers"
                  className="bg-white dark:bg-zinc-800/50"
                  required
                />
              </div>

              <div className="mb-4">
                <label
                  htmlFor="projectDescription"
                  aria-label="Project description"
                  className="block text-sm font-medium text-zinc-900 dark:text-zinc-300 mb-1"
                >
                  Project Description
                </label>
                <Input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="The design project!"
                  className="bg-white dark:bg-zinc-800/50"
                  required
                />
              </div>

              <div>
                {/* biome-ignore lint/a11y/noLabelWithoutControl: TODO: Fix me */}
                <label className="block text-sm font-medium text-zinc-900 dark:text-zinc-300 mb-3">
                  Project Icon
                </label>
                <div className="relative">
                  <div className="grid grid-cols-8 gap-2 max-h-[240px] overflow-y-auto p-2 rounded-lg border border-zinc-200 dark:border-zinc-700/50">
                    {Object.entries(icons).map(([name, Icon]) => (
                      <button
                        key={name}
                        type="button"
                        onClick={() => setSelectedIcon(name)}
                        className={cn(
                          "p-2 rounded-lg transition-colors flex items-center justify-center group",
                          selectedIcon === name
                            ? "bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400"
                            : "text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800",
                        )}
                        title={name}
                      >
                        <Icon className="w-5 h-5" />
                      </button>
                    ))}
                  </div>
                  <div className="absolute left-0 right-0 bottom-0 h-8 bg-gradient-to-t from-white dark:from-zinc-900 to-transparent pointer-events-none" />
                </div>
                <div className="flex items-center gap-2 mt-2 px-2">
                  <IconComponent className="w-4 h-4 text-zinc-400" />
                  <span className="text-sm text-zinc-500 dark:text-zinc-400">
                    {selectedIcon}
                  </span>
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <Dialog.Close asChild>
                  <Button
                    type="button"
                    className="bg-zinc-100 text-zinc-900 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-100 dark:hover:bg-zinc-700"
                  >
                    Cancel
                  </Button>
                </Dialog.Close>
                <Button
                  type="submit"
                  className="bg-indigo-600 text-white hover:bg-indigo-500 dark:bg-indigo-500 dark:hover:bg-indigo-400"
                >
                  Create Project
                </Button>
              </div>
            </form>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export default CreateProjectModal;
