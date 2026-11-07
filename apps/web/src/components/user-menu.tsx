import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
  type ReactNode,
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/components/providers/auth-provider/hooks/use-auth";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useSignOut from "@/hooks/mutations/use-sign-out";
import useGetConfig from "@/hooks/queries/config/use-get-config";
import { useLocale } from "@/hooks/use-locale";
import { cn } from "@/lib/cn";
import { toast } from "@/lib/toast";
import useProjectStore from "@/store/project";
import { useUserPreferencesStore } from "@/store/user-preferences";

const ACCOUNT_URL = "https://sso.facile.studio/account";
const HELP_URL = "https://facile.studio/docs/opus";
const LOCALES = [
  ["fr-FR", "Français"],
  ["en-US", "English"],
  ["es-ES", "Español"],
  ["de-DE", "Deutsch"],
] as const;
const THEMES = ["system", "light", "dark"] as const;

type View = "root" | "language" | "theme";
type Row = {
  id: string;
  label: string;
  detail?: string;
  href?: string;
  checked?: boolean;
  kind: "item" | "radio" | "submenu" | "back";
  tone?: "danger";
  run?: () => void;
};

const line =
  "flex min-h-11 w-full cursor-pointer items-center gap-2.5 rounded-fc-md px-3 text-left text-fc-sm focus-visible:bg-fc-surface focus-visible:outline-none hover:bg-fc-surface";

export function UserMenu() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const { data: config } = useGetConfig();
  const { mutateAsync: signOut } = useSignOut(config?.customOAuthLogoutUrl);
  const queryClient = useQueryClient();
  const { setProject } = useProjectStore();
  const navigate = useNavigate();
  const { locale, setLocale } = useLocale();
  const { theme, setTheme } = useUserPreferencesStore();
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<View>("root");
  const root = useRef<HTMLDivElement>(null);
  const menu = useRef<HTMLDivElement>(null);
  const menuId = useId();

  const close = useCallback((restoreFocus = true) => {
    setOpen(false);
    setView("root");
    if (restoreFocus) root.current?.querySelector("button")?.focus();
  }, []);

  useEffect(() => {
    if (!open) return;
    const away = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) close(false);
    };
    document.addEventListener("pointerdown", away);
    return () => document.removeEventListener("pointerdown", away);
  }, [open, close]);

  const mountMenu = useCallback((node: HTMLDivElement | null) => {
    menu.current = node;
    node
      ?.querySelector<HTMLElement>('[role^="menuitem"][tabindex="0"]')
      ?.focus();
  }, []);

  if (!user) return null;

  const name = user.name || t("navigation:userMenu.unnamedUser");
  const initials = user.name
    ? user.name
        .split(" ")
        .map((part) => part[0])
        .join("")
        .toUpperCase()
    : user.email?.substring(0, 2).toUpperCase() || "U";

  const handleSignOut = async () => {
    try {
      await signOut();
      queryClient.clear();
      setProject(undefined);
      toast.success(t("navigation:userMenu.signedOutSuccess"));
    } catch (error) {
      toast.error(
        error instanceof Error
          ? error.message
          : t("navigation:userMenu.signOutFailed"),
      );
    }
  };

  const themeLabel = (value: (typeof THEMES)[number]) =>
    t(`navigation:userMenu.theme_${value}`);
  const localeLabel = LOCALES.find(([code]) => code === locale)?.[1] ?? locale;

  const rows = ((): Row[] => {
    if (view === "language") {
      return [
        {
          id: "back",
          kind: "back",
          label: t("navigation:userMenu.language"),
        },
        ...LOCALES.map(([code, label]) => ({
          id: code,
          kind: "radio" as const,
          label,
          checked: code === locale,
          run: () => void setLocale(code),
        })),
      ];
    }
    if (view === "theme") {
      return [
        { id: "back", kind: "back", label: t("navigation:userMenu.theme") },
        ...THEMES.map((value) => ({
          id: value,
          kind: "radio" as const,
          label: themeLabel(value),
          checked: value === theme,
          run: () => setTheme(value),
        })),
      ];
    }
    return [
      {
        id: "account",
        kind: "item",
        label: t("navigation:userMenu.account"),
        href: ACCOUNT_URL,
      },
      {
        id: "settings",
        kind: "item",
        label: t("navigation:userMenu.settings"),
        run: () => navigate({ to: "/dashboard/settings/account/information" }),
      },
      {
        id: "language",
        kind: "submenu",
        label: t("navigation:userMenu.language"),
        detail: localeLabel,
      },
      {
        id: "theme",
        kind: "submenu",
        label: t("navigation:userMenu.theme"),
        detail: themeLabel(theme),
      },
      {
        id: "help",
        kind: "item",
        label: t("navigation:userMenu.help"),
        href: HELP_URL,
      },
      {
        id: "signout",
        kind: "item",
        label: t("navigation:userMenu.logOut"),
        tone: "danger",
        run: () => void handleSignOut(),
      },
    ];
  })();

  const choose = (row: Row) => {
    if (row.kind === "submenu") return setView(row.id as View);
    if (row.kind === "back") return setView("root");
    row.run?.();
    if (row.kind === "item") close(!row.href);
  };

  const keydown = (event: React.KeyboardEvent) => {
    const items = Array.from(
      menu.current?.querySelectorAll<HTMLElement>('[role^="menuitem"]') ?? [],
    );
    const at = items.indexOf(document.activeElement as HTMLElement);
    const focus = (index: number) => {
      for (const [i, el] of items.entries()) {
        el.tabIndex = i === index ? 0 : -1;
      }
      items[index]?.focus();
    };
    const row = rows[at];
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      if (view !== "root") return setView("root");
      return close();
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      focus((at + 1) % items.length);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      focus((at - 1 + items.length) % items.length);
    } else if (event.key === "Home") {
      event.preventDefault();
      focus(0);
    } else if (event.key === "End") {
      event.preventDefault();
      focus(items.length - 1);
    } else if (event.key === "ArrowRight" && row?.kind === "submenu") {
      event.preventDefault();
      setView(row.id as View);
    } else if (event.key === "ArrowLeft" && view !== "root") {
      event.preventDefault();
      setView("root");
    } else if (event.key === "Tab") {
      close(false);
    }
  };

  const content = (row: Row): ReactNode => (
    <>
      <span
        className={cn(
          "min-w-0 flex-1 truncate",
          row.checked && "font-semibold",
        )}
      >
        {row.label}
      </span>
      {row.detail && (
        <span className="max-w-28 shrink-0 truncate text-fc-xs text-fc-fg-muted">
          {row.detail}
        </span>
      )}
      {row.kind === "submenu" && <span aria-hidden="true">›</span>}
      {row.checked && <span aria-hidden="true">✓</span>}
      {row.href && (
        <span className="sr-only">{t("navigation:userMenu.opensNewTab")}</span>
      )}
    </>
  );

  return (
    <div ref={root} className="relative w-full">
      <button
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label={t("navigation:userMenu.open", { name })}
        onClick={() => (open ? close() : setOpen(true))}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            setOpen(true);
          }
        }}
        className={cn(
          "flex min-h-12 w-full cursor-pointer items-center gap-2.5 overflow-hidden rounded-fc-md px-2.5 py-1.5 text-left text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring",
          open && "bg-fc-surface",
        )}
      >
        <Avatar className="size-7 shrink-0 text-fc-xs">
          <AvatarImage src={user.image ?? ""} alt="" />
          <AvatarFallback className="text-fc-xs font-medium">
            {initials}
          </AvatarFallback>
        </Avatar>
        <span className="flex min-w-0 flex-1 flex-col leading-tight">
          <span className="truncate text-fc-sm font-medium">{name}</span>
          {user.email && (
            <span className="truncate text-fc-xs text-fc-fg-muted">
              {user.email}
            </span>
          )}
        </span>
      </button>
      {open && (
        <div
          key={view}
          ref={mountMenu}
          id={menuId}
          role="menu"
          aria-label={name}
          tabIndex={-1}
          onKeyDown={keydown}
          className="absolute bottom-full left-0 z-50 mb-2 max-h-[70dvh] min-w-60 overflow-y-auto overscroll-contain rounded-fc-md bg-fc-bg p-1 text-fc-fg shadow-lg"
        >
          {rows.map((row, index) => {
            const tone =
              row.tone === "danger" ? "text-fc-danger" : "text-fc-fg";
            const role = row.kind === "radio" ? "menuitemradio" : "menuitem";
            if (row.href) {
              return (
                <a
                  key={row.id}
                  role={role}
                  tabIndex={index === 0 ? 0 : -1}
                  href={row.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={cn(line, tone)}
                  onClick={() => choose(row)}
                >
                  {content(row)}
                </a>
              );
            }
            return (
              <button
                key={row.id}
                type="button"
                role={role}
                {...(row.kind === "radio"
                  ? { "aria-checked": row.checked }
                  : {})}
                aria-haspopup={row.kind === "submenu" ? "menu" : undefined}
                tabIndex={index === 0 ? 0 : -1}
                className={cn(line, tone, row.kind === "radio" && "pl-9")}
                onClick={() => choose(row)}
              >
                {content(row)}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
