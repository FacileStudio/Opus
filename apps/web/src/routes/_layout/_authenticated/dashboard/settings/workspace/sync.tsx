import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import triggerNookPoolSync from "@/fetchers/nook-pool/trigger-sync";
import useDeleteNookPool from "@/hooks/mutations/workspace/use-delete-nook-pool";
import useSaveNookPool from "@/hooks/mutations/workspace/use-save-nook-pool";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import useNookPoolConfig from "@/hooks/queries/workspace/use-nook-pool-config";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace/sync",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { data: workspace } = useActiveWorkspace();
  const { data: config, isLoading } = useNookPoolConfig(workspace?.id);
  const { mutateAsync: saveConfig, isPending: isSaving } = useSaveNookPool();
  const { mutateAsync: deleteConfig, isPending: isDeleting } =
    useDeleteNookPool();

  const [instanceUrl, setInstanceUrl] = useState("");
  const [secret, setSecret] = useState("");
  const [urlTouched, setUrlTouched] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncResult, setSyncResult] = useState<{
    projects_synced: number;
    tasks_synced: number;
  } | null>(null);

  const isConnected = !!config?.connected;
  const isFromEnv = !!config?.fromEnv;

  const handleConnect = useCallback(async () => {
    if (!workspace?.id) return;
    const url = (instanceUrl || config?.instanceUrl || "").trim();
    const sec = (secret || "").trim();

    if (!url || !sec) {
      toast.error("Instance URL and secret are required");
      return;
    }

    try {
      await saveConfig({
        workspaceId: workspace.id,
        instanceUrl: url,
        secret: sec,
        enabled: true,
      });
      toast.success("Connected to Nook Pool");
      setSecret("");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to connect");
    }
  }, [workspace?.id, instanceUrl, secret, saveConfig, config]);

  const handleDisconnect = useCallback(async () => {
    if (!workspace?.id) return;

    try {
      await deleteConfig({ workspaceId: workspace.id });
      toast.success("Disconnected from Nook Pool");
      setInstanceUrl("");
      setSecret("");
      setUrlTouched(false);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to disconnect",
      );
    }
  }, [workspace?.id, deleteConfig]);

  const handleSync = useCallback(async () => {
    if (!workspace?.id) return;
    setIsSyncing(true);
    setSyncResult(null);
    try {
      const result = await triggerNookPoolSync(workspace.id);
      if (result.success) {
        setSyncResult({
          projects_synced: result.projects_synced ?? 0,
          tasks_synced: result.tasks_synced ?? 0,
        });
        toast.success(
          `Synced ${result.projects_synced} projects and ${result.tasks_synced} tasks`,
        );
      } else {
        toast.error(result.error || "Sync failed");
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Sync failed");
    } finally {
      setIsSyncing(false);
    }
  }, [workspace?.id]);

  const displayUrl =
    urlTouched || !config ? instanceUrl : config?.instanceUrl || instanceUrl;

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48" />
          <div className="h-4 bg-muted rounded w-96" />
        </div>
      </div>
    );
  }

  return (
    <>
      <PageTitle title="Sync" />
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">Nook Pool</h1>
          <p className="text-muted-foreground">
            Connect this workspace to a Nook Pool instance to sync projects and
            tasks with other apps.
          </p>
        </div>

        <div className="space-y-6">
          <div className="space-y-1">
            <h2 className="text-md font-medium">Connection</h2>
            <p className="text-xs text-muted-foreground">
              Configure the Nook Pool instance for this workspace.
            </p>
          </div>

          <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Status</p>
                <p className="text-xs text-muted-foreground">
                  Current connection status
                </p>
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={`inline-block size-2.5 rounded-full ${
                    isConnected ? "bg-green-500" : "bg-muted-foreground/40"
                  }`}
                />
                <span className="text-sm">
                  {isConnected ? "Connected" : "Not connected"}
                </span>
              </div>
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Instance URL</p>
                <p className="text-xs text-muted-foreground">
                  The URL of your Nook Pool server
                </p>
                {isFromEnv && !urlTouched && (
                  <p className="text-xs text-blue-500">
                    Pre-filled from environment variable
                  </p>
                )}
              </div>
              <Input
                className="w-64"
                placeholder="https://nook.example.com"
                value={displayUrl}
                onChange={(e) => {
                  setInstanceUrl(e.target.value);
                  setUrlTouched(true);
                }}
                disabled={isConnected}
              />
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-sm font-medium">Secret</p>
                <p className="text-xs text-muted-foreground">
                  Shared secret for authenticating with the Pool instance
                </p>
              </div>
              <Input
                className="w-64"
                type="password"
                placeholder={isConnected ? "••••••••" : "Enter secret"}
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                disabled={isConnected}
              />
            </div>

            <Separator />

            <div className="flex justify-end">
              {isConnected ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDisconnect}
                  disabled={isDeleting}
                  className="text-destructive hover:text-destructive"
                >
                  {isDeleting ? "Disconnecting..." : "Disconnect"}
                </Button>
              ) : (
                <Button size="sm" onClick={handleConnect} disabled={isSaving}>
                  {isSaving ? "Connecting..." : "Connect"}
                </Button>
              )}
            </div>
          </div>
        </div>

        {isConnected && (
          <>
            <div className="space-y-6">
              <div className="space-y-1">
                <h2 className="text-md font-medium">Synced events</h2>
                <p className="text-xs text-muted-foreground">
                  Events that are synced between this workspace and other apps.
                </p>
              </div>

              <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <p className="text-sm font-medium">Emitting</p>
                    <ul className="space-y-1 text-xs text-muted-foreground">
                      <li>project.created</li>
                      <li>project.updated</li>
                      <li>project.deleted</li>
                      <li>task.created</li>
                      <li>task.updated</li>
                      <li>task.deleted</li>
                    </ul>
                  </div>
                  <div className="space-y-2">
                    <p className="text-sm font-medium">Listening</p>
                    <ul className="space-y-1 text-xs text-muted-foreground">
                      <li>project.created</li>
                      <li>project.updated</li>
                      <li>project.deleted</li>
                      <li>task.created</li>
                      <li>task.updated</li>
                      <li>task.deleted</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <div className="space-y-1">
                <h2 className="text-md font-medium">Initial sync</h2>
                <p className="text-xs text-muted-foreground">
                  Push all existing projects and tasks to the Pool so other
                  connected apps can see them.
                </p>
              </div>

              <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium">Sync existing data</p>
                    <p className="text-xs text-muted-foreground">
                      Safe to run multiple times — duplicates are automatically
                      skipped.
                    </p>
                  </div>
                  <Button size="sm" onClick={handleSync} disabled={isSyncing}>
                    {isSyncing ? "Syncing..." : "Sync all"}
                  </Button>
                </div>
                {syncResult && (
                  <p className="text-xs text-muted-foreground">
                    Synced {syncResult.projects_synced} projects and{" "}
                    {syncResult.tasks_synced} tasks.
                  </p>
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
}
