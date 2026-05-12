import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createWorkspaceWebhook } from "@/fetchers/workspace-webhook/create-workspace-webhook";
import { deleteWorkspaceWebhook } from "@/fetchers/workspace-webhook/delete-workspace-webhook";
import { testWorkspaceWebhook } from "@/fetchers/workspace-webhook/test-workspace-webhook";
import { updateWorkspaceWebhook } from "@/fetchers/workspace-webhook/update-workspace-webhook";

export function useCreateWorkspaceWebhook(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name: string;
      url: string;
      secret?: string;
      events?: string;
      enabled?: boolean;
    }) => createWorkspaceWebhook(workspaceId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["workspace-webhooks", workspaceId],
      });
    },
  });
}

export function useUpdateWorkspaceWebhook(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      webhookId: string;
      name?: string;
      url?: string;
      secret?: string | null;
      events?: string;
      enabled?: boolean;
    }) => {
      const { webhookId, ...rest } = data;
      return updateWorkspaceWebhook(workspaceId, webhookId, rest);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["workspace-webhooks", workspaceId],
      });
    },
  });
}

export function useDeleteWorkspaceWebhook(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (webhookId: string) =>
      deleteWorkspaceWebhook(workspaceId, webhookId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["workspace-webhooks", workspaceId],
      });
    },
  });
}

export function useTestWorkspaceWebhook(workspaceId: string) {
  return useMutation({
    mutationFn: (webhookId: string) =>
      testWorkspaceWebhook(workspaceId, webhookId),
  });
}
