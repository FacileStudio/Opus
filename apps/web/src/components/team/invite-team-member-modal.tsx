import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import {
  CheckCircle as Check,
  Copy,
  Link,
  CloseCircle as X,
} from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod/v4";
import useInviteWorkspaceUser from "@/hooks/mutations/workspace-user/use-invite-workspace-user";
import { authClient } from "@/lib/auth-client";
import { toast } from "@/lib/toast";
import { Route } from "@/routes/_layout/_authenticated/dashboard/workspace/$workspaceId/members";
import { Button } from "../ui/button";
import { Dialog, DialogClose, DialogPopup, DialogTitle } from "../ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Input } from "../ui/input";

type Props = {
  open: boolean;
  onClose: () => void;
};

const teamMemberSchema = z.object({
  email: z.string(),
});

type TeamMemberFormValues = z.infer<typeof teamMemberSchema>;

function InviteTeamMemberModal({ open, onClose }: Props) {
  const { t } = useTranslation();
  const { mutateAsync, isPending } = useInviteWorkspaceUser();
  const queryClient = useQueryClient();
  const { workspaceId } = Route.useParams();
  const [inviteLink, setInviteLink] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const form = useForm<TeamMemberFormValues>({
    resolver: standardSchemaResolver(teamMemberSchema),
    defaultValues: {
      email: "",
    },
  });

  const buildInviteLink = (invitationId: string) =>
    `${window.location.origin}/invitation/accept/${invitationId}`;

  const onSubmit = async ({ email }: TeamMemberFormValues) => {
    try {
      const data = await mutateAsync({ email, workspaceId, role: "member" });
      await queryClient.refetchQueries({
        queryKey: ["workspace-users", workspaceId],
      });

      let invitationId = data?.id;

      if (!invitationId) {
        const { data: invitations } =
          await authClient.organization.listInvitations({
            query: { organizationId: workspaceId },
          });
        const match = invitations?.find(
          (inv) =>
            inv.email.toLowerCase() === email.toLowerCase() &&
            inv.status === "pending",
        );
        invitationId = match?.id;
      }

      if (invitationId) {
        setInviteLink(buildInviteLink(invitationId));
      } else {
        toast.success(t("team:inviteModal.success"));
        resetAndCloseModal();
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : t("team:inviteModal.error"),
      );
    }
  };

  const handleCopyLink = async () => {
    if (!inviteLink) return;
    await navigator.clipboard.writeText(inviteLink);
    setCopied(true);
    toast.success(t("team:inviteModal.linkCopied"));
    setTimeout(() => setCopied(false), 2000);
  };

  const resetInviteTeamMember = async () => {
    await queryClient.invalidateQueries({
      queryKey: ["workspace-users", workspaceId],
    });
    form.reset();
    setInviteLink(null);
    setCopied(false);
  };

  const resetAndCloseModal = () => {
    resetInviteTeamMember();
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={resetAndCloseModal}>
      <DialogPopup className="w-full max-w-md">
        <div className="bg-card rounded-lg shadow-xl">
          <div className="flex items-center justify-between p-4 border-b border-border">
            <DialogTitle className="text-lg font-semibold text-foreground">
              {t("team:inviteModal.title")}
            </DialogTitle>
            <DialogClose
              className="text-muted-foreground hover:text-foreground"
              render={<button type="button" />}
            >
              <X size={20} />
            </DialogClose>
          </div>

          {inviteLink ? (
            <div className="p-4 space-y-4">
              <p className="text-sm text-muted-foreground">
                {t("team:inviteModal.success")}
              </p>
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-2 flex-1 rounded-md border border-border bg-muted/50 px-3 py-2 text-sm text-muted-foreground overflow-hidden">
                  <Link className="h-4 w-4 shrink-0" />
                  <span className="truncate">{inviteLink}</span>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleCopyLink}
                  className="shrink-0"
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-green-500" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <div className="flex justify-end">
                <Button type="button" onClick={resetAndCloseModal}>
                  {t("common:actions.close")}
                </Button>
              </div>
            </div>
          ) : (
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="p-4">
                <div className="space-y-4">
                  <div>
                    <FormField
                      control={form.control}
                      name="email"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel className="block text-sm font-medium text-foreground mb-1">
                            {t("team:inviteModal.emailLabel")}
                          </FormLabel>
                          <FormControl>
                            <Input
                              {...field}
                              placeholder={t(
                                "team:inviteModal.emailPlaceholder",
                              )}
                              className="bg-card/50"
                              autoFocus
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </div>

                <div className="flex justify-end gap-2 mt-6">
                  <DialogClose
                    render={
                      <Button
                        className="bg-muted text-foreground hover:bg-accent"
                        type="button"
                      />
                    }
                  >
                    {t("common:actions.cancel")}
                  </DialogClose>
                  <Button type="submit" disabled={isPending}>
                    {t("team:inviteModal.sendInvitation")}
                  </Button>
                </div>
              </form>
            </Form>
          )}
        </div>
      </DialogPopup>
    </Dialog>
  );
}

export default InviteTeamMemberModal;
