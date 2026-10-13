import { standardSchemaResolver } from "@hookform/resolvers/standard-schema";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import PageTitle from "@/components/page-title";
import useAuth from "@/components/providers/auth-provider/hooks/use-auth";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { UserColor } from "@/constants/user-colors";
import {
  normalizeUserColor,
  USER_COLOR_LABELS,
  USER_COLORS,
} from "@/constants/user-colors";
import useUpdateUserProfile from "@/hooks/mutations/use-update-user-profile";
import { cn } from "@/lib/cn";
import { toast } from "@/lib/toast";

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/settings/account/information",
)({
  component: RouteComponent,
});

type ProfileFormValues = {
  name: string;
  email: string;
};

type NormalizedProfileValues = {
  name: string;
  email: string;
};

function normalizeProfileValues(
  data: ProfileFormValues,
): NormalizedProfileValues {
  return {
    name: data.name.trim(),
    email: data.email,
  };
}

function RouteComponent() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const { mutateAsync: updateProfile } = useUpdateUserProfile();
  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isSavingRef = useRef(false);
  const queuedSaveRef = useRef<ProfileFormValues | null>(null);
  const lastSavedRef = useRef<NormalizedProfileValues | null>(null);
  const profileSchema = z.object({
    name: z
      .string()
      .min(1, t("settings:informationPage.validation.nameRequired"))
      .min(2, t("settings:informationPage.validation.nameShort")),
    email: z
      .string()
      .email(t("settings:informationPage.validation.invalidEmail")),
  });

  const [selectedColor, setSelectedColor] = useState<UserColor>(() =>
    normalizeUserColor(
      (user as Record<string, unknown> | undefined)?.color as
        | string
        | undefined,
    ),
  );

  useEffect(() => {
    if (user) {
      setSelectedColor(
        normalizeUserColor(
          (user as Record<string, unknown>)?.color as string | undefined,
        ),
      );
    }
  }, [user]);

  const handleColorChange = useCallback(
    async (color: UserColor) => {
      setSelectedColor(color);
      try {
        await updateProfile({ color: color.replace("#", "") });
        await queryClient.invalidateQueries({ queryKey: ["session"] });
      } catch {
        toast.error("Failed to update color");
      }
    },
    [updateProfile, queryClient],
  );

  const profileForm = useForm<ProfileFormValues>({
    resolver: standardSchemaResolver(profileSchema),
    mode: "onChange",
    defaultValues: {
      name: user?.name || "",
      email: user?.email || "",
    },
  });

  useEffect(() => {
    if (!user) return;

    const nextValues = {
      name: user.name || "",
      email: user.email || "",
    };
    lastSavedRef.current = normalizeProfileValues(nextValues);

    if (!profileForm.formState.isDirty) {
      profileForm.reset(nextValues);
    }
  }, [user, profileForm]);

  const saveProfile = useCallback(
    async (data: ProfileFormValues) => {
      const normalizedData = normalizeProfileValues(data);

      if (lastSavedRef.current?.name === normalizedData.name) {
        return;
      }

      if (isSavingRef.current) {
        queuedSaveRef.current = data;
        return;
      }

      isSavingRef.current = true;

      try {
        await updateProfile({
          name: normalizedData.name,
        });

        profileForm.reset(normalizedData, { keepDirty: false });
        lastSavedRef.current = normalizedData;
        queuedSaveRef.current = null;

        await queryClient.invalidateQueries({ queryKey: ["session"] });
        toast.success(t("settings:informationPage.updateSuccess"));
      } catch (error) {
        toast.error(
          error instanceof Error
            ? error.message
            : t("settings:informationPage.updateError"),
        );
      } finally {
        isSavingRef.current = false;

        if (queuedSaveRef.current) {
          const queuedData = queuedSaveRef.current;
          queuedSaveRef.current = null;
          await saveProfile(queuedData);
        }
      }
    },
    [t, updateProfile, queryClient, profileForm],
  );

  const debouncedSave = useCallback(
    (data: ProfileFormValues) => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }

      debounceTimeoutRef.current = setTimeout(() => {
        saveProfile(data);
      }, 1000);
    },
    [saveProfile],
  );

  useEffect(() => {
    const subscription = profileForm.watch(() => {
      if (profileForm.formState.isDirty && profileForm.formState.isValid) {
        debouncedSave(profileForm.getValues());
      }
    });

    return () => subscription.unsubscribe();
  }, [profileForm, debouncedSave]);

  useEffect(() => {
    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
    };
  }, []);

  return (
    <>
      <PageTitle title={t("settings:informationPage.pageTitle")} />
      <div className="mx-auto w-full max-w-3xl space-y-8">
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">
            {t("settings:informationPage.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("settings:informationPage.subtitle")}
          </p>
        </div>

        <section className="space-y-5">
          <div>
            <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">
              {t("settings:informationPage.sectionTitle")}
            </p>
            <h2 className="font-semibold">
              {t("settings:informationPage.sectionSubtitle")}
            </h2>
          </div>

          <div className="flex items-center gap-4">
            <Avatar className="h-12 w-12">
              <AvatarImage src={user?.image ?? ""} alt={user?.name || ""} />
              <AvatarFallback
                className="text-sm font-semibold text-white"
                style={{ backgroundColor: selectedColor }}
              >
                {user?.name?.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="text-sm font-medium">{user?.name}</p>
              <p className="text-xs text-muted-foreground">{user?.email}</p>
              {user?.image && (
                <p className="text-xs text-muted-foreground">
                  {t("settings:informationPage.syncedFromSSO")}
                </p>
              )}
            </div>
          </div>

          <Form {...profileForm}>
            <form className="space-y-4">
              <FormField
                control={profileForm.control}
                name="name"
                render={({ field }) => (
                  <FormItem className="grid gap-1.5">
                    <FormLabel className="text-sm text-muted-foreground">
                      {t("settings:informationPage.fullName")}
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t(
                          "settings:informationPage.fullNamePlaceholder",
                        )}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={profileForm.control}
                name="email"
                render={({ field }) => (
                  <FormItem className="grid gap-1.5">
                    <FormLabel className="text-sm text-muted-foreground">
                      {t("settings:informationPage.email")}
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t(
                          "settings:informationPage.emailPlaceholder",
                        )}
                        {...field}
                        disabled
                        value={user?.email || ""}
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
              Personalization
            </p>
            <h2 className="font-semibold">Color</h2>
          </div>

          <div className="flex flex-wrap gap-2">
            {USER_COLORS.map((color) => (
              <button
                key={color}
                type="button"
                onClick={() => handleColorChange(color)}
                className={cn(
                  "flex items-center gap-2 rounded-full border px-3 py-2 text-sm transition-colors",
                  selectedColor === color
                    ? "border-foreground bg-foreground text-background"
                    : "border-border bg-background hover:border-foreground/30 hover:bg-muted/40",
                )}
                aria-pressed={selectedColor === color}
              >
                <span
                  className="inline-block h-3 w-3 shrink-0 rounded-full border border-black/5"
                  style={{ backgroundColor: color }}
                />
                {USER_COLOR_LABELS[color]}
              </button>
            ))}
          </div>
        </section>
      </div>
    </>
  );
}
