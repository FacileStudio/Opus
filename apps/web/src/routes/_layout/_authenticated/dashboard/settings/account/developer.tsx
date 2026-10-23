import { Key, TrashBinTrash } from "@solar-icons/react";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
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
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { type ApiToken, authClient } from "@/lib/auth-client";
import { formatDateMedium } from "@/lib/format";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/account/developer",
)({
  component: RouteComponent,
});

function formatDay(value: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return formatDateMedium(date);
}

function RouteComponent() {
  const { t } = useTranslation();
  const [tokens, setTokens] = useState<ApiToken[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [name, setName] = useState("");
  const [nameError, setNameError] = useState("");
  const [created, setCreated] = useState<ApiToken | null>(null);
  const [copied, setCopied] = useState(false);
  const [tokenToDelete, setTokenToDelete] = useState<ApiToken | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const refresh = useCallback(async () => {
    const { data, error } = await authClient.apiTokens.list();
    if (error) {
      toast.error(error.message);
      setTokens([]);
      return;
    }
    setTokens(data ?? []);
  }, []);

  useEffect(() => {
    refresh().finally(() => setIsLoading(false));
  }, [refresh]);

  const handleCreate = useCallback(async () => {
    const trimmed = name.trim();
    if (!trimmed) {
      setNameError(t("settings:apiKey.createDialog.validation.nameRequired"));
      return;
    }
    if (trimmed.length < 3) {
      setNameError(t("settings:apiKey.createDialog.validation.nameShort"));
      return;
    }

    setIsCreating(true);
    const { data, error } = await authClient.apiTokens.create({
      name: trimmed,
    });
    setIsCreating(false);

    if (error || !data) {
      toast.error(
        error?.message ?? t("settings:apiKey.createDialog.failedCreate"),
      );
      return;
    }

    setIsCreateOpen(false);
    setName("");
    setNameError("");
    setCopied(false);
    setCreated(data);
    await refresh();
  }, [name, refresh, t]);

  const handleDelete = useCallback(async () => {
    if (!tokenToDelete) return;
    setIsDeleting(true);
    const { error } = await authClient.apiTokens.revoke({
      id: tokenToDelete.id,
    });
    setIsDeleting(false);
    setTokenToDelete(null);

    if (error) {
      toast.error(error.message);
      return;
    }
    toast.success(t("settings:apiKey.table.toastDeleted"));
    await refresh();
  }, [refresh, t, tokenToDelete]);

  const handleCopy = useCallback(
    async (value: string) => {
      try {
        await navigator.clipboard.writeText(value);
        setCopied(true);
        toast.success(t("settings:apiKey.createdModal.toastCopied"));
      } catch {
        toast.error("Could not copy to the clipboard");
      }
    },
    [t],
  );

  const token = created?.token ?? "";
  const configSnippet = `# ~/.opus.yml\napi_key: "${token}"`;
  const envSnippet = `export OPUS_TOKEN="${token}"`;

  return (
    <>
      <PageTitle title={t("settings:developerPage.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">
            {t("settings:developerPage.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("settings:developerPage.subtitle")}
          </p>
        </div>

        <Card>
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div className="space-y-1">
              <CardTitle>
                {t("settings:developerPage.apiKeysCardTitle")}
              </CardTitle>
              <CardDescription>
                {t("settings:developerPage.apiKeysCardDescription")}
              </CardDescription>
            </div>
            <Button
              size="sm"
              className="shrink-0"
              onClick={() => setIsCreateOpen(true)}
            >
              <Key className="h-4 w-4" />
              {t("settings:developerPage.createApiKey")}
            </Button>
          </CardHeader>

          <div className="px-4 pb-4">
            {isLoading ? (
              <p className="py-6 text-sm text-muted-foreground">
                {t("settings:apiKey.table.loading")}
              </p>
            ) : tokens.length === 0 ? (
              <p className="py-6 text-sm text-muted-foreground">
                {t("settings:apiKey.table.empty")}
              </p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      {t("settings:apiKey.table.columnName")}
                    </TableHead>
                    <TableHead>
                      {t("settings:apiKey.table.columnCreated")}
                    </TableHead>
                    <TableHead>
                      {t("settings:apiKey.table.columnLastUsed")}
                    </TableHead>
                    <TableHead className="text-right">
                      {t("settings:apiKey.table.columnActions")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">
                        {item.name || t("settings:apiKey.table.unnamedKey")}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDay(item.createdAt)}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDay(item.lastUsedAt)}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("settings:apiKey.table.deleteAria", {
                            name:
                              item.name ||
                              t("settings:apiKey.table.deleteAriaFallback"),
                          })}
                          onClick={() => setTokenToDelete(item)}
                        >
                          <TrashBinTrash className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </Card>
      </div>

      <Dialog
        open={isCreateOpen}
        onOpenChange={(open) => {
          setIsCreateOpen(open);
          if (!open) {
            setName("");
            setNameError("");
          }
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t("settings:apiKey.createDialog.title")}</DialogTitle>
            <DialogDescription>
              {t("settings:apiKey.createDialog.description")}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2 px-1">
            <Label htmlFor="token-name">
              {t("settings:apiKey.createDialog.nameLabel")}
            </Label>
            <Input
              id="token-name"
              value={name}
              onChange={(event) => {
                setName(event.target.value);
                setNameError("");
              }}
              placeholder={t("settings:apiKey.createDialog.namePlaceholder")}
            />
            {nameError ? (
              <p className="text-xs text-destructive">{nameError}</p>
            ) : (
              <p className="text-xs text-muted-foreground">
                {t("settings:apiKey.createDialog.nameDescription")}
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsCreateOpen(false)}
            >
              {t("common:actions.cancel")}
            </Button>
            <Button size="sm" onClick={handleCreate} disabled={isCreating}>
              {isCreating
                ? t("settings:apiKey.createDialog.creating")
                : t("settings:apiKey.createDialog.create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={created !== null}
        onOpenChange={(open) => !open && setCreated(null)}
      >
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("settings:apiKey.createdModal.title")}</DialogTitle>
            <DialogDescription>
              {t("settings:apiKey.createdModal.alertDescription")}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 px-1">
            <div className="space-y-2">
              <Label>{t("settings:apiKey.createdModal.yourApiKey")}</Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 overflow-x-auto rounded-md bg-sidebar px-3 py-2 font-mono text-xs">
                  {token}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  className="shrink-0"
                  onClick={() => handleCopy(token)}
                >
                  {copied
                    ? t("settings:apiKey.createdModal.copied")
                    : t("settings:apiKey.createdModal.copy")}
                </Button>
              </div>
            </div>

            <div className="space-y-2">
              <Label>{t("settings:apiKey.createdModal.configTitle")}</Label>
              <pre className="overflow-x-auto rounded-md bg-sidebar px-3 py-2 font-mono text-xs">
                {configSnippet}
              </pre>
              <pre className="overflow-x-auto rounded-md bg-sidebar px-3 py-2 font-mono text-xs">
                {envSnippet}
              </pre>
            </div>
          </div>

          <DialogFooter>
            <Button size="sm" onClick={() => setCreated(null)}>
              {t("settings:apiKey.createdModal.done")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={tokenToDelete !== null}
        onOpenChange={(open) => !open && setTokenToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("settings:apiKey.table.deleteConfirmTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("settings:apiKey.table.deleteConfirmDescription", {
                name:
                  tokenToDelete?.name ||
                  t("settings:apiKey.table.deleteFallbackName"),
              })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogClose>
              <Button variant="outline" size="sm">
                {t("common:actions.cancel")}
              </Button>
            </AlertDialogClose>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting
                ? t("settings:apiKey.table.deleting")
                : t("settings:apiKey.table.delete")}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
