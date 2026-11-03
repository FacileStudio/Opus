import { SquareArrowRightUp } from "@solar-icons/react";
import { useTranslation } from "react-i18next";
import { buttonVariants } from "@/components/ui/button";
import useGetConfig from "@/hooks/queries/config/use-get-config";
import { cn } from "@/lib/cn";

/**
 * Registre's space management page, or an empty string while Opus still
 * manages workspaces itself.
 */
export function useRegistreSpacesUrl(): string {
  const { data } = useGetConfig();
  return (
    (data as { registreSpacesUrl?: string } | undefined)?.registreSpacesUrl ??
    ""
  );
}

/**
 * Links to the Registre page where a linked workspace's name and members are
 * managed, or to the space list when no space id is given. It renders nothing
 * until the server reports where Registre is.
 */
export function ManageInRegistreButton({
  spaceId,
  className,
}: {
  spaceId?: string | null;
  className?: string;
}) {
  const { t } = useTranslation();
  const spacesUrl = useRegistreSpacesUrl();
  if (!spacesUrl) {
    return null;
  }
  const href = spaceId
    ? `${spacesUrl}/${encodeURIComponent(spaceId)}`
    : spacesUrl;
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        buttonVariants({ variant: "outline", size: "xs" }),
        className,
      )}
    >
      <SquareArrowRightUp className="w-3 h-3" />
      {t("workspace:registre.manage")}
    </a>
  );
}

/**
 * Explains that workspaces are created in Registre and links there.
 */
export function RegistreCreateNotice() {
  const { t } = useTranslation();
  return (
    <div className="space-y-4 text-center">
      <p className="text-sm text-muted-foreground">
        {t("workspace:registre.createHint")}
      </p>
      <ManageInRegistreButton className="gap-1" />
    </div>
  );
}
