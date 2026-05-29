import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import PageTitle from "@/components/page-title";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import getPoolEvents, {
  type PoolEventToggle,
} from "@/fetchers/nook-pool/get-pool-events";
import triggerNookPoolSync from "@/fetchers/nook-pool/trigger-sync";
import updatePoolEvents from "@/fetchers/nook-pool/update-pool-events";
import useDeleteNookPool from "@/hooks/mutations/workspace/use-delete-nook-pool";
import useSaveNookPool from "@/hooks/mutations/workspace/use-save-nook-pool";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import useNookPoolConfig from "@/hooks/queries/workspace/use-nook-pool-config";
import { toast } from "@/lib/toast";

const EVENT_LABELS: Record<string, string> = {
  "task.created": "Task created",
  "task.updated": "Task updated",
  "task.deleted": "Task deleted",
  "task.status_changed": "Task status changed",
  "task.actor_changed": "Task actor changed",
  "task.priority_changed": "Task priority changed",
  "project.created": "Project created",
  "project.updated": "Project updated",
  "project.deleted": "Project deleted",
};

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
  const [poolEventToggles, setPoolEventToggles] = useState<PoolEventToggle[]>(
    [],
  );

  const isConnected = !!config?.connected;
  const isFromEnv = !!config?.fromEnv;

  useEffect(() => {
    if (!workspace?.id || !isConnected) return;
    getPoolEvents(workspace.id)
      .then(setPoolEventToggles)
      .catch(() => {});
  }, [workspace?.id, isConnected]);

  const handleToggleEvent = useCallback(
    async (event: string, enabled: boolean) => {
      if (!workspace?.id) return;
      const prev = poolEventToggles;
      const next = poolEventToggles.map((e) =>
        e.event === event ? { ...e, enabled } : e,
      );
      setPoolEventToggles(next);
      try {
        const result = await updatePoolEvents(workspace.id, next);
        setPoolEventToggles(result);
      } catch {
        setPoolEventToggles(prev);
        toast.error("Failed to update event toggle");
      }
    },
    [workspace?.id, poolEventToggles],
  );

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
      <div className="max-w-xl py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-muted rounded w-32" />
          <div className="h-4 bg-muted rounded w-80" />
        </div>
      </div>
    );
  }

  return (
    <>
      <PageTitle title="Sync" />
      <div className="max-w-xl space-y-8">
        <div>
          <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
            Sync
          </p>
          <h1 className="text-lg font-semibold">Nook Pool</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Connect this workspace to a Nook Pool instance to sync projects and
            tasks with other apps.
          </p>
        </div>

        <section className="space-y-5">
          <div>
            <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
              Connection
            </p>
            <h2 className="font-semibold">Pool instance</h2>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Status</span>
              <div className="flex items-center gap-2">
                <span
                  className={`inline-block size-2 rounded-full ${
                    isConnected ? "bg-green-500" : "bg-muted-foreground/40"
                  }`}
                />
                <span className="text-sm">
                  {isConnected ? "Connected" : "Not connected"}
                </span>
              </div>
            </div>

            <div className="grid gap-1.5">
              <span className="text-sm text-muted-foreground">
                Instance URL
              </span>
              <Input
                placeholder="https://nook.example.com"
                value={displayUrl}
                onChange={(e) => {
                  setInstanceUrl(e.target.value);
                  setUrlTouched(true);
                }}
                disabled={isConnected}
              />
              {isFromEnv && !urlTouched && (
                <p className="text-xs text-blue-500">
                  Pre-filled from environment variable
                </p>
              )}
            </div>

            <div className="grid gap-1.5">
              <span className="text-sm text-muted-foreground">Secret</span>
              <Input
                type="password"
                placeholder={isConnected ? "••••••••" : "Enter secret"}
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                disabled={isConnected}
              />
            </div>

            <div className="flex justify-end pt-1">
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
        </section>

        {isConnected && (
          <>
            <section className="space-y-5">
              <div>
                <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
                  Events
                </p>
                <h2 className="font-semibold">Pool events</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  Choose which events are sent to Nook. Nook handles routing
                  from there.
                </p>
              </div>

              <div className="divide-y divide-border">
                {poolEventToggles.map((evt) => (
                  <div
                    key={evt.event}
                    className="flex items-center justify-between py-3 first:pt-0 last:pb-0"
                  >
                    <div className="flex flex-col gap-0.5">
                      <span className="text-sm font-medium">
                        {EVENT_LABELS[evt.event] ?? evt.event}
                      </span>
                      <span className="font-mono text-xs text-muted-foreground">
                        {evt.event}
                      </span>
                    </div>
                    <button
                      type="button"
                      role="switch"
                      aria-checked={evt.enabled}
                      aria-label={`Toggle ${evt.event}`}
                      className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors ${evt.enabled ? "bg-primary" : "bg-muted"}`}
                      onClick={() => handleToggleEvent(evt.event, !evt.enabled)}
                    >
                      <span
                        className={`pointer-events-none inline-block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform ${evt.enabled ? "translate-x-5" : "translate-x-0"}`}
                      />
                    </button>
                  </div>
                ))}
              </div>
            </section>

            <section className="space-y-5">
              <div>
                <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
                  Data
                </p>
                <h2 className="font-semibold">Initial sync</h2>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <p className="text-sm text-muted-foreground">
                    Push all existing projects and tasks to the Pool. Safe to
                    run multiple times.
                  </p>
                  <Button
                    size="sm"
                    className="shrink-0 ml-4"
                    onClick={handleSync}
                    disabled={isSyncing}
                  >
                    {isSyncing ? "Syncing..." : "Sync all"}
                  </Button>
                </div>
                {syncResult && (
                  <p className="text-sm text-muted-foreground">
                    Synced {syncResult.projects_synced} projects and{" "}
                    {syncResult.tasks_synced} tasks.
                  </p>
                )}
              </div>
            </section>
          </>
        )}
      </div>
    </>
  );
}
