# Opus — Deployment

How the image is built, how the single container is wired, and what to set when deploying to la
ruche through Dokploy or to Kubernetes through the bundled Helm chart.

## The image

`Dockerfile` is a four-stage build, all stages based on `oven/bun` with pnpm 10.32.1 installed
globally:

1. `api-builder` — installs the full workspace from the lockfile, builds `packages/email` and then
   bundles `apps/api` with esbuild into `apps/api/dist`.
2. `api-deps` — reinstalls with `--prod --frozen-lockfile --no-optional` (with the `prepare` script
   neutered so Husky does not run), then copies in the built API, the Drizzle migrations and the
   built email package.
3. `web-builder` — installs, copies `i18n/` and `apps/web`, and runs the Vite production build.
4. `runtime` — `oven/bun:1-slim` plus nginx and wget. Runs as the unprivileged `appuser` (uid
   1001), exposes `80`, and declares a `HEALTHCHECK` against `http://127.0.0.1:80/api/health`.

`apps/api/drizzle` is copied into the runtime stage. Migrations ship with the image and run at
startup — there is no separate migration job.

## One container, two processes

This is the one-container / one-router / one-hostname shape the suite standardizes on. nginx owns
the public port; the API is never exposed directly.

```
Traefik ──▶ :80 nginx ──┬──▶ /api/*                    ──▶ 127.0.0.1:1337  Hono API
                         ├──▶ /.well-known/oauth-*      ──▶ 127.0.0.1:1337
                         ├──▶ /.well-known/*            ──▶ 404
                         └──▶ /*                        ──▶ /usr/share/nginx/html, SPA fallback
```

`deploy/opus-entrypoint.sh` is the container command and, in order:

1. Derives `OPUS_API_URL` as `<OPUS_CLIENT_URL>/api` when it is unset.
2. Derives `DATABASE_URL` from the `POSTGRES_*` variables, URL-encoding user and password. **Exits
   1** when neither `DATABASE_URL` nor `POSTGRES_PASSWORD` is set.
3. Generates a random `AUTH_SECRET` when unset and logs a warning — sessions will not survive a
   restart, so always set it explicitly in production.
4. Runs `/docker-entrypoint.d/env.sh`, which rewrites the `OPUS_API_URL` and `OPUS_CLIENT_URL`
   placeholders inside the built JavaScript and injects the MCP OAuth discovery JSON into the nginx
   config.
5. Starts the API, polls `/api/health` until it answers, then starts nginx.
6. Watches both processes; if either dies it kills the other and exits with the first non-zero
   status.

nginx sets `X-Content-Type-Options`, `X-Frame-Options: SAMEORIGIN`, `Referrer-Policy` and a
`client_max_body_size` of 25 MB, and gzips text responses.

## Docker Compose

```sh
cp .env.sample .env
docker compose up -d
```

Two services:

- `postgres` — `postgres:16-alpine`, named container `opus-db`, data on the `opus-db-data` volume,
  `pg_isready` healthcheck.
- `opus` — built from the local `Dockerfile` and tagged `ghcr.io/facilestudio/opus:latest`, named
  container `opus-app`, depends on Postgres being healthy, mounts `opus-uploads` at
  `/app/apps/api/data/uploads` for workspace logos, restarts unless stopped, and repeats the
  image's healthcheck.

Neither service publishes a host port. The `opus` container is reachable only on the compose
network, which is what you want behind Traefik and what you must change for a direct local run.

A commented-out `redis` service (Valkey) is included for the HA setup described below.

## Behind Traefik on la ruche

Route one hostname to the `opus` container's port `80`. Nothing else in the container should be
exposed. The variables that must match the public hostname:

```sh
OPUS_CLIENT_URL=https://opus.facile.studio
```

`OPUS_API_URL` is then derived as `https://opus.facile.studio/api`, `better-auth` picks up its
`baseURL` and trusted origins from those two, and CORS falls back to `OPUS_CLIENT_URL`. Setting
`CORS_ORIGINS` explicitly is still a good idea in production — leaving both unset makes the API
echo back any origin.

Because the client and the API share a hostname, cookies are issued `SameSite=Lax` without
`Secure`. Split them across subdomains and `better-auth` switches to
`SameSite=None; Secure; Partitioned` automatically. Use `COOKIE_DOMAIN` only if you actually need
a shared parent domain.

Prefer the `dokploy` CLI over SSH plus docker for managing the deployment.

## Scaling out

The default WebSocket broadcast adapter is in-memory, so more than one replica silently loses
real-time updates between them. Set `REDIS_URL` (or the sentinel or cluster variables) before
raising the replica count. Uploads have the same problem in reverse: workspace logos are written to
local disk, so either keep a single replica or back `UPLOAD_DIR` with shared storage.

## Kubernetes

`charts/opus` is a self-contained chart: Deployment, Service, ServiceAccount, optional PVC,
optional HPA, and both an Ingress and a Gateway API HTTPRoute. `postgresql.enabled` defaults to
`true` and deploys `postgres:16-alpine` inside the release with an 8 Gi PVC; point it at an
external database for anything serious.

```sh
helm install opus ./charts/opus --set opus.env.OPUS_CLIENT_URL=https://opus.example.com
```

Two things to check before installing:

- `opus.service.port` and `opus.service.targetPort` default to `5173`, but the image listens on
  `80`. Override both, or the Service will point at a closed port.
- The default Postgres credentials in `values.yaml` are `opus_user` / `opus_password`. Use
  `postgresql.auth.existingSecret`.

`opus.extraEnv` takes arbitrary environment variables for the application container.

## Health and observability

`GET /api/health` returns `{"status":"ok"}` as soon as the Hono app is listening. It says nothing
about the database, the pool connection or whether nginx is up — a green health check is not proof
that the front end is being served. Verify the root path returns the SPA too.

Startup logs are worth reading on a first deploy: the entrypoint announces every value it derived,
the migrator prints its progress, and the pool client logs a line per workspace connection.
