import { logger } from "@bogeychan/elysia-logger";
import { cors } from "@elysiajs/cors";
import { Elysia } from "elysia";
import project from "./project";
import task from "./task";
import user from "./user";
import { validateSessionToken } from "./user/controllers/validate-session-token";
import workspace from "./workspace";

const app = new Elysia()
  .state("userId", "")
  .use(cors())
  .use(logger())
  .use(user)
  .guard({
    async beforeHandle({ store, cookie: { session } }) {
      if (!session?.value) {
        return { user: null };
      }

      const { user, session: validatedSession } = await validateSessionToken(
        session.value,
      );

      if (!user || !validatedSession) {
        return { user: null };
      }

      store.userId = user.id;
    },
  })
  .get("/me", async ({ cookie: { session } }) => {
    const { user } = await validateSessionToken(session.value ?? "");

    if (user === null) {
      return { user: null };
    }

    return { user };
  })
  .use(workspace)
  .use(project)
  .use(task)
  .onError(({ code, error }) => {
    switch (code) {
      case "VALIDATION":
        return error.all;
      default:
        if (error instanceof Error) {
          return {
            name: error.name,
            message: error.message,
          };
        }
    }
  })
  .listen(1337);

export type App = typeof app;

console.log(
  `🦊 Elysia is running at ${app.server?.hostname}:${app.server?.port}`,
);
