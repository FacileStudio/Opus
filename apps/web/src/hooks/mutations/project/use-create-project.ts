import createProject from "@/fetchers/project/create-project";
import { useMutation } from "@tanstack/react-query";

function useCreateProject({
  name,
  description,
  workspaceId,
}: { name: string; description: string; workspaceId: string }) {
  return useMutation({
    mutationFn: () => createProject({ name, description, workspaceId }),
  });
}

export default useCreateProject;
