import { useTranslation } from "react-i18next";
import { IdeasIcon } from "@/components/ideas/ideas-icon";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { isSafeHref } from "@/lib/ideas";

const EXAMPLES = ["mail", "task", "follow_up"] as const;

export function IdeasUpsell({
  upgradeUrl,
}: {
  upgradeUrl: string | undefined;
}) {
  const { t } = useTranslation("ideas");

  return (
    <div className="flex flex-col gap-6" data-testid="ideas-upsell">
      <Card className="items-start gap-4 p-6">
        <IdeasIcon size={24} className="text-muted-foreground" />
        <h2 className="m-0 text-base font-medium">{t("upsell.title")}</h2>
        <p className="max-w-prose text-sm text-muted-foreground">
          {t("upsell.description")}
        </p>
        {isSafeHref(upgradeUrl) && (
          <a
            href={upgradeUrl}
            target="_blank"
            rel="noopener noreferrer"
            className={buttonVariants({
              size: "lg",
              className: "border-transparent",
            })}
          >
            {t("upsell.cta")}
          </a>
        )}
      </Card>
      <ul
        className="grid gap-4 md:grid-cols-3"
        aria-label={t("upsell.examples")}
      >
        {EXAMPLES.map((kind) => (
          <li key={kind} className="min-w-0">
            <Card className="h-full gap-3 bg-muted p-5">
              <Badge variant="secondary" className="self-start">
                {t("upsell.example")}
              </Badge>
              <h3 className="m-0 break-words text-base font-medium">
                {t(`upsell.${kind}.title`)}
              </h3>
              <p className="break-words text-sm text-muted-foreground">
                {t(`upsell.${kind}.why`)}
              </p>
            </Card>
          </li>
        ))}
      </ul>
    </div>
  );
}
