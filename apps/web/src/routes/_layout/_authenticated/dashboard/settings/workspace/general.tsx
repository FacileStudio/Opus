import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import { CloseCircle, Gallery } from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { getApiUrl } from "@/fetchers/get-api-url";
import {
  deleteWorkspaceLogo,
  uploadWorkspaceLogo,
} from "@/fetchers/workspace/upload-workspace-logo";
import useDeleteWorkspace from "@/hooks/mutations/workspace/use-delete-workspace";
import useUpdateWorkspace from "@/hooks/mutations/workspace/use-update-workspace";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/workspace/general",
)({
  component: RouteComponent,
});

type WorkspaceFormValues = {
  name: string;
  description?: string;
};

type NormalizedWorkspaceValues = {
  name: string;
  description: string;
};

function normalizeWorkspaceValues(
  data: WorkspaceFormValues,
): NormalizedWorkspaceValues {
  return {
    name: data.name.trim(),
    description: (data.description ?? "").trim(),
  };
}

/** Better Auth persists description as an organization additional field (DB column), not only inside metadata. */
function getWorkspaceDescription(
  workspace:
    | { description?: string | null; metadata?: unknown }
    | null
    | undefined,
): string {
  if (!workspace) return "";
  if (typeof workspace.description === "string") {
    return workspace.description;
  }
  if (
    typeof workspace.metadata === "object" &&
    workspace.metadata &&
    "description" in workspace.metadata
  ) {
    return String(
      (workspace.metadata as { description?: unknown }).description ?? "",
    );
  }
  return "";
}

function RouteComponent() {
  const { t } = useTranslation();
  const workspaceSchema = useMemo(
    () =>
      z.object({
        name: z
          .string()
          .min(1, t("settings:workspaceGeneral.validation.nameRequired"))
          .min(2, t("settings:workspaceGeneral.validation.nameShort")),
        description: z.string().optional(),
      }),
    [t],
  );

  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isSavingRef = useRef(false);
  const queuedSaveRef = useRef<WorkspaceFormValues | null>(null);
  const lastSavedRef = useRef<NormalizedWorkspaceValues | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isUploadingLogo, setIsUploadingLogo] = useState(false);
  const [logoVersion, setLogoVersion] = useState(Date.now);
  const logoInputRef = useRef<HTMLInputElement>(null);

  const { data: workspace } = useActiveWorkspace();
  const { mutateAsync: updateWorkspace } = useUpdateWorkspace();
  const { mutateAsync: deleteWorkspace, isPending: isDeleting } =
    useDeleteWorkspace();
  const workspaceDescription = getWorkspaceDescription(workspace);

  const workspaceForm = useForm<WorkspaceFormValues>({
    resolver: standardSchemaResolver(workspaceSchema),
    mode: "onChange",
    defaultValues: {
      name: workspace?.name || "",
      description: workspaceDescription,
    },
  });

  useEffect(() => {
    if (!workspace) return;

    const nextValues = {
      name: workspace.name || "",
      description: workspaceDescription,
    };
    lastSavedRef.current = normalizeWorkspaceValues(nextValues);

    if (!workspaceForm.formState.isDirty) {
      workspaceForm.reset(nextValues);
    }
  }, [workspace, workspaceDescription, workspaceForm]);

  const saveWorkspace = useCallback(
    async (data: WorkspaceFormValues) => {
      if (!workspace?.id) return;

      const normalizedData = normalizeWorkspaceValues(data);
      const nameChanged = lastSavedRef.current?.name !== normalizedData.name;
      const descriptionChanged =
        lastSavedRef.current?.description !== normalizedData.description;
      const hasChanges = nameChanged || descriptionChanged;

      if (!hasChanges) return;

      if (isSavingRef.current) {
        queuedSaveRef.current = data;
        return;
      }

      isSavingRef.current = true;

      try {
        const updatePayload: {
          workspaceId: string;
          name?: string;
          description?: string;
        } = {
          workspaceId: workspace.id,
        };

        if (nameChanged) {
          updatePayload.name = normalizedData.name;
        }

        if (descriptionChanged) {
          updatePayload.description = normalizedData.description;
        }

        await updateWorkspace(updatePayload);

        workspaceForm.reset(normalizedData, { keepDirty: false });
        lastSavedRef.current = normalizedData;
        queuedSaveRef.current = null;

        await queryClient.invalidateQueries({
          queryKey: ["active-organization"],
        });
        toast.success(t("settings:workspaceGeneral.toastUpdated"));
      } catch (error) {
        toast.error(
          error instanceof Error
            ? error.message
            : t("settings:workspaceGeneral.toastUpdateError"),
        );
      } finally {
        isSavingRef.current = false;

        if (queuedSaveRef.current) {
          const queuedData = queuedSaveRef.current;
          queuedSaveRef.current = null;
          await saveWorkspace(queuedData);
        }
      }
    },
    [workspace, updateWorkspace, queryClient, workspaceForm, t],
  );

  const handleDeleteWorkspace = useCallback(async () => {
    if (!workspace?.id) return;

    try {
      await deleteWorkspace({ workspaceId: workspace.id });
      toast.success(t("settings:workspaceGeneral.toastDeleted"));

      // Invalidate all workspace-related queries
      await queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      await queryClient.invalidateQueries({
        queryKey: ["active-organization"],
      });

      navigate({ to: "/dashboard" });
    } catch (error) {
      toast.error(
        error instanceof Error
          ? error.message
          : t("settings:workspaceGeneral.toastDeleteError"),
      );
    }
  }, [workspace?.id, deleteWorkspace, queryClient, navigate, t]);

  const debouncedSave = useCallback(
    (data: WorkspaceFormValues) => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }

      debounceTimeoutRef.current = setTimeout(() => {
        saveWorkspace(data);
      }, 1000);
    },
    [saveWorkspace],
  );

  useEffect(() => {
    const subscription = workspaceForm.watch(() => {
      if (workspaceForm.formState.isDirty && workspaceForm.formState.isValid) {
        debouncedSave(workspaceForm.getValues());
      }
    });

    return () => subscription.unsubscribe();
  }, [workspaceForm, debouncedSave]);

  useEffect(() => {
    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
    };
  }, []);

  return (
    <>
      <PageTitle title={t("settings:workspaceGeneral.pageTitle")} />
      <div className="max-w-xl space-y-8">
        <div>
          <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
            {t("navigation:page.settingsWorkspaceTab")}
          </p>
          <h1 className="text-lg font-semibold">
            {t("settings:workspaceGeneral.title")}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("settings:workspaceGeneral.subtitle")}
          </p>
        </div>

        <section className="space-y-5">
          <div>
            <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
              {t("settings:workspaceGeneral.workspaceInfoTitle")}
            </p>
            <h2 className="font-semibold">
              {t("settings:workspaceGeneral.workspaceInfoSubtitle")}
            </h2>
          </div>

          <Form {...workspaceForm}>
            <form className="space-y-4">
              <FormField
                control={workspaceForm.control}
                name="name"
                render={({ field }) => (
                  <FormItem className="grid gap-1.5">
                    <FormLabel className="text-sm text-muted-foreground">
                      {t("settings:workspaceGeneral.nameLabel")}
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t(
                          "settings:workspaceGeneral.namePlaceholder",
                        )}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={workspaceForm.control}
                name="description"
                render={({ field }) => (
                  <FormItem className="grid gap-1.5">
                    <FormLabel className="text-sm text-muted-foreground">
                      {t("settings:workspaceGeneral.descriptionLabel")}
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t(
                          "settings:workspaceGeneral.descriptionPlaceholder",
                        )}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </form>
          </Form>
        </section>

        <section className="space-y-5">
          <div>
            <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
              {t("settings:workspaceGeneral.logoTitle")}
            </p>
            <h2 className="font-semibold">
              {t("settings:workspaceGeneral.logoSubtitle")}
            </h2>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Avatar className="size-14 rounded-lg">
                {workspace?.logo ? (
                  <AvatarImage
                    src={getApiUrl(
                      `/workspace/${workspace.id}/logo?v=${logoVersion}`,
                    )}
                    alt={workspace.name}
                  />
                ) : null}
                <AvatarFallback className="rounded-lg bg-muted text-lg">
                  {workspace?.name?.charAt(0)?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <p className="text-sm text-muted-foreground">
                {t("settings:workspaceGeneral.logoHint")}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {workspace?.logo && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-muted-foreground"
                  disabled={isUploadingLogo}
                  onClick={async () => {
                    if (!workspace?.id) return;
                    try {
                      await deleteWorkspaceLogo(workspace.id);
                      await updateWorkspace({
                        workspaceId: workspace.id,
                        logo: "",
                      });
                      setLogoVersion(Date.now());
                      await queryClient.invalidateQueries({
                        queryKey: ["active-organization"],
                      });
                      toast.success(t("settings:workspaceGeneral.logoDeleted"));
                    } catch {
                      toast.error(
                        t("settings:workspaceGeneral.logoDeleteError"),
                      );
                    }
                  }}
                >
                  <CloseCircle className="size-4" />
                  {t("common:actions.remove")}
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                disabled={isUploadingLogo}
                onClick={() => logoInputRef.current?.click()}
              >
                <Gallery className="size-4" />
                {isUploadingLogo
                  ? t("settings:workspaceGeneral.logoUploading")
                  : t("settings:workspaceGeneral.logoUpload")}
              </Button>
              <input
                ref={logoInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp,image/svg+xml,image/gif"
                className="hidden"
                onChange={async (e) => {
                  const file = e.target.files?.[0];
                  if (!file || !workspace?.id) return;

                  setIsUploadingLogo(true);
                  try {
                    const result = await uploadWorkspaceLogo(
                      workspace.id,
                      file,
                    );
                    await updateWorkspace({
                      workspaceId: workspace.id,
                      logo: result.url,
                    });
                    setLogoVersion(Date.now());
                    await queryClient.invalidateQueries({
                      queryKey: ["active-organization"],
                    });
                    toast.success(t("settings:workspaceGeneral.logoUpdated"));
                  } catch (error) {
                    toast.error(
                      error instanceof Error
                        ? error.message
                        : t("settings:workspaceGeneral.logoUploadError"),
                    );
                  } finally {
                    setIsUploadingLogo(false);
                    if (logoInputRef.current) {
                      logoInputRef.current.value = "";
                    }
                  }
                }}
              />
            </div>
          </div>
        </section>

        <section className="space-y-5">
          <div>
            <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
              {t("settings:workspaceGeneral.dangerZone")}
            </p>
            <h2 className="font-semibold">
              {t("settings:workspaceGeneral.dangerZoneSubtitle")}
            </h2>
          </div>

          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              {t("settings:workspaceGeneral.deleteWorkspaceDescription")}
            </p>
            <Button
              variant="ghost"
              size="sm"
              className="shrink-0 ml-4 text-destructive hover:text-destructive transition-colors"
              type="button"
              onClick={() => setIsDeleteModalOpen(true)}
            >
              {t("settings:workspaceGeneral.deleteWorkspace")}
            </Button>
          </div>
        </section>

        <AlertDialog
          open={isDeleteModalOpen}
          onOpenChange={setIsDeleteModalOpen}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {t("settings:workspaceGeneral.deleteModalTitle")}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {t("settings:workspaceGeneral.deleteModalDescription", {
                  name: workspace?.name ?? "",
                })}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogClose>
                <Button variant="outline" size="sm">
                  {t("common:actions.cancel")}
                </Button>
              </AlertDialogClose>
              <AlertDialogClose
                onClick={handleDeleteWorkspace}
                disabled={isDeleting}
              >
                <Button variant="destructive" size="sm" disabled={isDeleting}>
                  {isDeleting
                    ? t("common:actions.deleting")
                    : t("settings:workspaceGeneral.deleteModalConfirm")}
                </Button>
              </AlertDialogClose>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </>
  );
}
