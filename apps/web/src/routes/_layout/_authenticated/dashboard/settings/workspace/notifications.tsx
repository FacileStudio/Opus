import {
  AddCircle,
  BellBing,
  CheckCircle,
  LinkMinimalistic,
  Pen,
  TestTube,
  TrashBinTrash,
} from "@solar-icons/react";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import type { WorkspaceWebhook } from "@/fetchers/workspace-webhook/get-workspace-webhooks";
import {
  useCreateWorkspaceWebhook,
  useDeleteWorkspaceWebhook,
  useTestWorkspaceWebhook,
  useUpdateWorkspaceWebhook,
} from "@/hooks/mutations/workspace-webhook/use-workspace-webhooks";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { useGetWorkspaceWebhooks } from "@/hooks/queries/workspace-webhook/use-get-workspace-webhooks";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace/notifications",
)({
  component: RouteComponent,
});

const AVAILABLE_EVENTS = [
  "task.created",
  "task.updated",
  "task.status_changed",
  "task.assignee_changed",
  "task.deleted",
  "comment.created",
  "member.joined",
  "member.removed",
] as const;

type WebhookFormData = {
  name: string;
  url: string;
  secret: string;
  events: string[];
  enabled: boolean;
};

const EMPTY_FORM: WebhookFormData = {
  name: "",
  url: "",
  secret: "",
  events: [],
  enabled: true,
};

function parseEvents(events: string): string[] {
  if (!events || events === "*") return [];
  return events.split(",").filter(Boolean);
}

function serializeEvents(events: string[]): string {
  if (events.length === 0) return "*";
  return events.join(",");
}

function EventCheckboxes({
  selected,
  onChange,
}: {
  selected: string[];
  onChange: (events: string[]) => void;
}) {
  const allSelected = selected.length === 0;

  return (
    <div className="space-y-2">
      <label className="flex items-center gap-2 text-xs">
        <Checkbox checked={allSelected} onCheckedChange={() => onChange([])} />
        <span className="text-muted-foreground">All events</span>
      </label>
      <div className="grid grid-cols-2 gap-1.5">
        {AVAILABLE_EVENTS.map((event) => (
          <label key={event} className="flex items-center gap-2 text-xs">
            <Checkbox
              checked={!allSelected && selected.includes(event)}
              disabled={allSelected}
              onCheckedChange={(checked) => {
                if (checked) {
                  onChange([...selected, event]);
                } else {
                  const next = selected.filter((e) => e !== event);
                  onChange(next.length === 0 ? [] : next);
                }
              }}
            />
            <span className="font-mono text-muted-foreground">{event}</span>
          </label>
        ))}
      </div>
    </div>
  );
}

function WebhookCard({
  webhook,
  workspaceId,
}: {
  webhook: WorkspaceWebhook;
  workspaceId: string;
}) {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = useState(false);
  const [form, setForm] = useState<WebhookFormData>({
    name: webhook.name,
    url: webhook.url,
    secret: "",
    events: parseEvents(webhook.events),
    enabled: webhook.enabled,
  });

  const updateMutation = useUpdateWorkspaceWebhook(workspaceId);
  const deleteMutation = useDeleteWorkspaceWebhook(workspaceId);
  const testMutation = useTestWorkspaceWebhook(workspaceId);

  const handleToggle = useCallback(
    (checked: boolean) => {
      updateMutation.mutate(
        { webhookId: webhook.id, enabled: checked },
        {
          onSuccess: () => {
            toast.success(t("settings:workspaceNotifications.toggleSuccess"));
          },
          onError: (err) => {
            toast.error(err.message);
          },
        },
      );
    },
    [updateMutation, webhook.id, t],
  );

  const handleSave = useCallback(() => {
    if (!form.name.trim() || !form.url.trim()) return;

    const payload: {
      webhookId: string;
      name?: string;
      url?: string;
      secret?: string | null;
      events?: string;
      enabled?: boolean;
    } = {
      webhookId: webhook.id,
      name: form.name.trim(),
      url: form.url.trim(),
      events: serializeEvents(form.events),
      enabled: form.enabled,
    };

    if (form.secret) {
      payload.secret = form.secret;
    }

    updateMutation.mutate(payload, {
      onSuccess: () => {
        setIsEditing(false);
        toast.success(t("settings:workspaceNotifications.updateSuccess"));
      },
      onError: (err) => {
        toast.error(err.message);
      },
    });
  }, [form, webhook.id, updateMutation, t]);

  const handleDelete = useCallback(() => {
    deleteMutation.mutate(webhook.id, {
      onSuccess: () => {
        toast.success(t("settings:workspaceNotifications.deleteSuccess"));
      },
      onError: (err) => {
        toast.error(err.message);
      },
    });
  }, [deleteMutation, webhook.id, t]);

  const handleTest = useCallback(() => {
    testMutation.mutate(webhook.id, {
      onSuccess: (result) => {
        if (result.success) {
          toast.success(
            t("settings:workspaceNotifications.testSuccess", {
              status: result.statusCode,
            }),
          );
        } else {
          toast.error(
            result.error || t("settings:workspaceNotifications.testFailed"),
          );
        }
      },
      onError: (err) => {
        toast.error(err.message);
      },
    });
  }, [testMutation, webhook.id, t]);

  const _eventDisplay =
    webhook.events === "*"
      ? t("settings:workspaceNotifications.allEvents")
      : webhook.events;

  if (isEditing) {
    return (
      <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
        <div className="space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium">
              {t("settings:workspaceNotifications.nameLabel")}
            </label>
            <Input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder={t("settings:workspaceNotifications.namePlaceholder")}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">
              {t("settings:workspaceNotifications.urlLabel")}
            </label>
            <Input
              value={form.url}
              onChange={(e) => setForm((f) => ({ ...f, url: e.target.value }))}
              placeholder="https://example.com/webhook"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">
              {t("settings:workspaceNotifications.secretLabel")}
            </label>
            <Input
              type="password"
              value={form.secret}
              onChange={(e) =>
                setForm((f) => ({ ...f, secret: e.target.value }))
              }
              placeholder={t(
                "settings:workspaceNotifications.secretPlaceholder",
              )}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium">
              {t("settings:workspaceNotifications.eventsLabel")}
            </label>
            <EventCheckboxes
              selected={form.events}
              onChange={(events) => setForm((f) => ({ ...f, events }))}
            />
          </div>
          <div className="flex items-center gap-2">
            <Switch
              checked={form.enabled}
              onCheckedChange={(checked) =>
                setForm((f) => ({ ...f, enabled: checked }))
              }
            />
            <span className="text-xs text-muted-foreground">
              {t("settings:workspaceNotifications.enabled")}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2 pt-2">
          <Button
            size="sm"
            onClick={handleSave}
            disabled={
              updateMutation.isPending || !form.name.trim() || !form.url.trim()
            }
          >
            {t("common:actions.save")}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setIsEditing(false);
              setForm({
                name: webhook.name,
                url: webhook.url,
                secret: "",
                events: parseEvents(webhook.events),
                enabled: webhook.enabled,
              });
            }}
          >
            {t("common:actions.cancel")}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-3 border border-border rounded-md p-4 bg-sidebar">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BellBing size={16} weight="Linear" />
          <span className="text-sm font-medium">{webhook.name}</span>
          {!webhook.enabled && (
            <Badge variant="outline" size="sm">
              {t("settings:workspaceNotifications.disabled")}
            </Badge>
          )}
        </div>
        <Switch checked={webhook.enabled} onCheckedChange={handleToggle} />
      </div>

      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <LinkMinimalistic size={14} weight="Linear" />
        <span className="truncate max-w-sm font-mono">{webhook.url}</span>
      </div>

      <div className="flex flex-wrap gap-1">
        {webhook.events === "*" ? (
          <Badge variant="secondary" size="sm">
            {t("settings:workspaceNotifications.allEvents")}
          </Badge>
        ) : (
          webhook.events.split(",").map((event) => (
            <Badge key={event} variant="outline" size="sm">
              {event}
            </Badge>
          ))
        )}
      </div>

      <Separator />

      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs"
          onClick={() => setIsEditing(true)}
        >
          <Pen size={14} weight="Linear" />
          {t("common:actions.edit")}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs"
          onClick={handleTest}
          disabled={testMutation.isPending}
        >
          <TestTube size={14} weight="Linear" />
          {t("settings:workspaceNotifications.test")}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-destructive hover:text-destructive"
          onClick={handleDelete}
          disabled={deleteMutation.isPending}
        >
          <TrashBinTrash size={14} weight="Linear" />
          {t("common:actions.delete")}
        </Button>
      </div>
    </div>
  );
}

function RouteComponent() {
  const { t } = useTranslation();
  const { data: workspace } = useActiveWorkspace();
  const workspaceId = workspace?.id ?? "";
  const { data: webhooks, isLoading } = useGetWorkspaceWebhooks(workspaceId);
  const createMutation = useCreateWorkspaceWebhook(workspaceId);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [form, setForm] = useState<WebhookFormData>(EMPTY_FORM);

  const handleCreate = useCallback(() => {
    if (!form.name.trim() || !form.url.trim()) return;

    createMutation.mutate(
      {
        name: form.name.trim(),
        url: form.url.trim(),
        secret: form.secret || undefined,
        events: serializeEvents(form.events),
        enabled: form.enabled,
      },
      {
        onSuccess: () => {
          setShowCreateForm(false);
          setForm(EMPTY_FORM);
          toast.success(t("settings:workspaceNotifications.createSuccess"));
        },
        onError: (err) => {
          toast.error(err.message);
        },
      },
    );
  }, [form, createMutation, t]);

  return (
    <>
      <PageTitle title={t("settings:workspaceNotifications.pageTitle")} />
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">
            {t("settings:workspaceNotifications.title")}
          </h1>
          <p className="text-muted-foreground">
            {t("settings:workspaceNotifications.subtitle")}
          </p>
        </div>

        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h2 className="text-md font-medium">
                {t("settings:workspaceNotifications.webhooksTitle")}
              </h2>
              <p className="text-xs text-muted-foreground">
                {t("settings:workspaceNotifications.webhooksSubtitle")}
              </p>
            </div>
            {!showCreateForm && (
              <Button
                size="sm"
                className="gap-1.5"
                onClick={() => setShowCreateForm(true)}
              >
                <AddCircle size={16} weight="Linear" />
                {t("settings:workspaceNotifications.addWebhook")}
              </Button>
            )}
          </div>

          {showCreateForm && (
            <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
              <div className="space-y-3">
                <div className="space-y-1">
                  <label className="text-xs font-medium">
                    {t("settings:workspaceNotifications.nameLabel")}
                  </label>
                  <Input
                    value={form.name}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, name: e.target.value }))
                    }
                    placeholder={t(
                      "settings:workspaceNotifications.namePlaceholder",
                    )}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium">
                    {t("settings:workspaceNotifications.urlLabel")}
                  </label>
                  <Input
                    value={form.url}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, url: e.target.value }))
                    }
                    placeholder="https://example.com/webhook"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium">
                    {t("settings:workspaceNotifications.secretLabel")}
                  </label>
                  <Input
                    type="password"
                    value={form.secret}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, secret: e.target.value }))
                    }
                    placeholder={t(
                      "settings:workspaceNotifications.secretPlaceholder",
                    )}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium">
                    {t("settings:workspaceNotifications.eventsLabel")}
                  </label>
                  <EventCheckboxes
                    selected={form.events}
                    onChange={(events) => setForm((f) => ({ ...f, events }))}
                  />
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={form.enabled}
                    onCheckedChange={(checked) =>
                      setForm((f) => ({ ...f, enabled: checked }))
                    }
                  />
                  <span className="text-xs text-muted-foreground">
                    {t("settings:workspaceNotifications.enabled")}
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-2 pt-2">
                <Button
                  size="sm"
                  onClick={handleCreate}
                  disabled={
                    createMutation.isPending ||
                    !form.name.trim() ||
                    !form.url.trim()
                  }
                >
                  <CheckCircle size={16} weight="Linear" />
                  {t("settings:workspaceNotifications.createWebhook")}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setShowCreateForm(false);
                    setForm(EMPTY_FORM);
                  }}
                >
                  {t("common:actions.cancel")}
                </Button>
              </div>
            </div>
          )}

          {isLoading && (
            <div className="text-sm text-muted-foreground py-8 text-center">
              {t("common:states.loading")}
            </div>
          )}

          {!isLoading &&
            webhooks &&
            webhooks.length === 0 &&
            !showCreateForm && (
              <div className="border border-dashed border-border rounded-md p-8 text-center space-y-2">
                <BellBing
                  size={32}
                  weight="Linear"
                  className="mx-auto text-muted-foreground"
                />
                <p className="text-sm text-muted-foreground">
                  {t("settings:workspaceNotifications.emptyState")}
                </p>
              </div>
            )}

          {webhooks && webhooks.length > 0 && (
            <div className="space-y-3">
              {webhooks.map((webhook) => (
                <WebhookCard
                  key={webhook.id}
                  webhook={webhook}
                  workspaceId={workspaceId}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
