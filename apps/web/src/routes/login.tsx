import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod/v4";

const loginSearchSchema = z.object({
  error: z.string().optional(),
});

export const Route = createFileRoute("/login")({
  validateSearch: loginSearchSchema,
  beforeLoad: ({ search }) => {
    throw redirect({
      to: "/auth/sign-in",
      search: { error: search.error },
    });
  },
});
