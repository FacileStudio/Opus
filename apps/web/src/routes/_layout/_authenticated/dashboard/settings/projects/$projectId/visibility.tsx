import {
  Copy,
  Link as LinkIcon,
  TrashBinMinimalistic,
} from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useParams } from "@tanstack/react-router";
import { useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import PageTitle from "@/components/page-title";
import {
  AlertDialog,
  AlertDialogClose,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import useUpdateProject from "@/hooks/mutations/project/use-update-project";
import useCreateShareLink from "@/hooks/mutations/share-link/use-create-share-link";
import useDeleteShareLink from "@/hooks/mutations/share-link/use-delete-share-link";
import useUpdateShareLink from "@/hooks/mutations/share-link/use-update-share-link";
import useGetProject from "@/hooks/queries/project/use-get-project";
import useGetShareLinks from "@/hooks/queries/share-link/use-get-share-links";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/projects/$projectId/visibility",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();
  const { projectId } = useParams({ strict: false });
  const { data: workspace } = useActiveWorkspace();
  const { data: project } = useGetProject({
    id: projectId || "",
    workspaceId: workspace?.id || "",
  });

  const queryClient = useQueryClient();
  const { mutateAsync: updateProject } = useUpdateProject();
  const savingRef = useRef(false);

  const handleToggle = useCallback(async () => {
    if (!project) return;
    if (savingRef.current) return;
    savingRef.current = true;
    try {
      await updateProject({
        id: project.id,
        name: project.name,
        slug: project.slug,
        description: project.description || "",
        icon: project.icon || "Layout",
        isPublic: !project.isPublic,
      });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["projects"] }),
        queryClient.invalidateQueries({
          queryKey: ["projects", workspace?.id],
        }),
        queryClient.invalidateQueries({
          queryKey: ["projects", workspace?.id, project.id],
        }),
      ]);
      toast.success(t("settings:projectVisibility.toastUpdated"));
    } catch (e) {
      toast.error(
        e instanceof Error
          ? e.message
          : t("settings:projectVisibility.toastUpdateError"),
      );
    } finally {
      savingRef.current = false;
    }
  }, [project, updateProject, queryClient, workspace?.id, t]);

  const origin = window.location.origin;

  const publicUrl = project?.id ? `${origin}/public-project/${project.id}` : "";

  return (
    <>
      <PageTitle title={t("settings:projectVisibility.pageTitle")} />
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">
            {t("settings:projectVisibility.title")}
          </h1>
          <p className="text-muted-foreground">
            {t("settings:projectVisibility.subtitle")}
          </p>
        </div>

        <div className="space-y-6">
          <div className="space-y-1">
            <h2 className="text-md font-medium">
              {t("settings:projectVisibility.sectionTitle")}
            </h2>
            <p className="text-xs text-muted-foreground">
              {t("settings:projectVisibility.sectionSubtitle")}
            </p>
          </div>

          <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-sm font-medium">
                  {t("settings:projectVisibility.publicAccess")}
                </Label>
                <p className="text-xs text-muted-foreground">
                  {t("settings:projectVisibility.publicAccessHint")}
                </p>
              </div>
              <Switch
                checked={!!project?.isPublic}
                onCheckedChange={handleToggle}
              />
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-sm font-medium">
                  {t("settings:projectVisibility.publicUrl")}
                </Label>
                <p className="text-xs text-muted-foreground">
                  {t("settings:projectVisibility.publicUrlHint")}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Input readOnly value={publicUrl} className="w-96" />
                <Button
                  size="sm"
                  onClick={() => {
                    if (!publicUrl) return;
                    navigator.clipboard
                      .writeText(publicUrl)
                      .then(() =>
                        toast.success(
                          t("settings:projectVisibility.copiedToast"),
                        ),
                      );
                  }}
                >
                  {t("settings:projectVisibility.copy")}
                </Button>
              </div>
            </div>
          </div>
        </div>

        {projectId && <ShareLinksSection projectId={projectId} />}
      </div>
    </>
  );
}

function ShareLinksSection({ projectId }: { projectId: string }) {
  const { t } = useTranslation();
  const origin = window.location.origin;
  const { data: links } = useGetShareLinks(projectId);
  const { mutateAsync: createLink, isPending: isCreating } =
    useCreateShareLink(projectId);
  const { mutateAsync: updateLink } = useUpdateShareLink(projectId);
  const { mutateAsync: deleteLink } = useDeleteShareLink(projectId);

  const [newLabel, setNewLabel] = useState("");
  const [newExpiry, setNewExpiry] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleCreate = async () => {
    try {
      await createLink({
        projectId,
        label: newLabel || undefined,
        expiresAt: newExpiry || undefined,
      });
      toast.success(t("settings:projectVisibility.shareLinks.createToast"));
      setNewLabel("");
      setNewExpiry("");
      setShowForm(false);
    } catch {
      toast.error(t("settings:projectVisibility.shareLinks.createError"));
    }
  };

  const handleToggleActive = async (id: string, isActive: boolean) => {
    try {
      await updateLink({ id, data: { isActive } });
    } catch {
      toast.error(t("settings:projectVisibility.shareLinks.updateError"));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteLink(id);
      toast.success(t("settings:projectVisibility.shareLinks.deleteToast"));
    } catch {
      toast.error(t("settings:projectVisibility.shareLinks.deleteError"));
    }
  };

  const copyLink = (token: string) => {
    navigator.clipboard
      .writeText(`${origin}/shared/${token}`)
      .then(() =>
        toast.success(t("settings:projectVisibility.shareLinks.copyToast")),
      )
      .catch(() =>
        toast.error(t("settings:projectVisibility.shareLinks.createError")),
      );
  };

  const getStatus = (link: { isActive: boolean; expiresAt: string | null }) => {
    if (!link.isActive) return "revoked";
    if (link.expiresAt && new Date(link.expiresAt) < new Date())
      return "expired";
    return "active";
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-md font-medium">
          {t("settings:projectVisibility.shareLinks.title")}
        </h2>
        <p className="text-xs text-muted-foreground">
          {t("settings:projectVisibility.shareLinks.subtitle")}
        </p>
      </div>

      <div className="space-y-4 border border-border rounded-md p-4 bg-sidebar">
        {!showForm ? (
          <Button size="sm" variant="outline" onClick={() => setShowForm(true)}>
            <LinkIcon className="h-3.5 w-3.5 mr-1.5" />
            {t("settings:projectVisibility.shareLinks.create")}
          </Button>
        ) : (
          <div className="space-y-3 border border-border rounded-md p-3 bg-background">
            <Input
              placeholder={t(
                "settings:projectVisibility.shareLinks.labelPlaceholder",
              )}
              value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
            />
            <div className="flex items-center gap-2">
              <Label className="text-xs text-muted-foreground shrink-0">
                {t("settings:projectVisibility.shareLinks.expiresAt")}
              </Label>
              <Input
                type="date"
                value={newExpiry}
                min={new Date().toISOString().split("T")[0]}
                onChange={(e) => setNewExpiry(e.target.value)}
                className="w-44"
              />
            </div>
            <div className="flex items-center gap-2">
              <Button size="sm" onClick={handleCreate} disabled={isCreating}>
                {t("settings:projectVisibility.shareLinks.create")}
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  setShowForm(false);
                  setNewLabel("");
                  setNewExpiry("");
                }}
              >
                {t("common:actions.cancel")}
              </Button>
            </div>
          </div>
        )}

        {links && links.length > 0 ? (
          <div className="space-y-3">
            {links.map((link) => {
              const status = getStatus(link);
              return (
                <div
                  key={link.id}
                  className="flex items-center justify-between gap-3 border border-border rounded-md p-3 bg-background"
                >
                  <div className="flex-1 min-w-0 space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">
                        {link.label || `${origin}/shared/${link.token}`}
                      </span>
                      <span
                        className={`px-1.5 py-0.5 text-[10px] rounded font-medium ${
                          status === "active"
                            ? "bg-green-500/10 text-green-600"
                            : status === "expired"
                              ? "bg-yellow-500/10 text-yellow-600"
                              : "bg-red-500/10 text-red-600"
                        }`}
                      >
                        {t(`settings:projectVisibility.shareLinks.${status}`)}
                      </span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span>
                        {link.accessCount > 0
                          ? t(
                              "settings:projectVisibility.shareLinks.accessCount",
                              { count: link.accessCount },
                            )
                          : t(
                              "settings:projectVisibility.shareLinks.neverAccessed",
                            )}
                      </span>
                      {link.expiresAt && (
                        <span>
                          {t("settings:projectVisibility.shareLinks.expiresAt")}{" "}
                          {formatDate(link.expiresAt)}
                        </span>
                      )}
                      {!link.expiresAt && (
                        <span>
                          {t("settings:projectVisibility.shareLinks.noExpiry")}
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => copyLink(link.token)}
                    >
                      <Copy className="h-3.5 w-3.5" />
                    </Button>
                    <Switch
                      checked={link.isActive}
                      onCheckedChange={(checked) =>
                        handleToggleActive(link.id, checked)
                      }
                    />
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setDeletingId(link.id)}
                    >
                      <TrashBinMinimalistic className="h-3.5 w-3.5 text-destructive" />
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          !showForm && (
            <p className="text-xs text-muted-foreground">
              {t("settings:projectVisibility.shareLinks.empty")}
            </p>
          )
        )}
      </div>

      <AlertDialog
        open={!!deletingId}
        onOpenChange={(open) => {
          if (!open) setDeletingId(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("common:actions.delete")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("settings:projectVisibility.shareLinks.deleteConfirm")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogClose>
              <Button variant="outline" size="sm">
                {t("common:actions.cancel")}
              </Button>
            </AlertDialogClose>
            <AlertDialogClose
              onClick={() => {
                const id = deletingId;
                if (id) handleDelete(id);
              }}
            >
              <Button variant="destructive" size="sm">
                {t("common:actions.delete")}
              </Button>
            </AlertDialogClose>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
