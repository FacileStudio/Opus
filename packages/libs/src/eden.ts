import { treaty } from "@elysiajs/eden";
import type { App } from "@kaneo/api";

export const api = treaty<App>(process.env.API_URL ?? "http://0.0.0.0:1337", {
  fetch: {
    credentials: "include",
  },
});
