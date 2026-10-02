# Opus — Deployment

How the image is built, how the single container is wired, and what to set when deploying to la
ruche through Dokploy or to Kubernetes through the bundled Helm chart.

## The image

`Dockerfile` is a three-stage build:

1. `web-builder` — `oven/bun:1` with pnpm 10.32.1 installed globally. Installs the workspace from
   the lockfile and builds the React SPA in `apps/web`, which reaches the Go binary as `/client`.
2. `go-builder` — `golang:1.26-alpine`. Downloads `apps/api`'s modules and compiles a single
   static binary (`CGO_ENABLED=0`, `-trimpath`, stripped).
3. `runtime` — `alpine:3.21` with `ca-certificates`, `tzdata` and `wget`. Runs as the
   unprivileged `appuser` (uid 1001), copies the SPA to `/client` and the binary to
   `/usr/local/bin/opus-api`, exposes `27538`, and declares a `HEALTHCHECK` against
   `http://127.0.0.1:27538/api/health`.

There is no Node, no nginx and no separate migration job: one Go process serves both the API and
the built client, and migrations run at startup.

## One process, one port

```
Traefik ──▶ :27538 opus-api ──┬──▶ /api/*     Go handlers (auth, workspaces, tasks, …)
                           ├──▶ /health, /ready, /api/health, /api/ready
                           ├──▶ /ws, /api/ws   realtime WebSocket hub
                           └──▶ /*      /client, SPA fallback to index.html
```

`deploy/opus-entrypoint.sh` is the container command and, in order:

1. Derives `OPUS_API_URL` as `<OPUS_CLIENT_URL>/api` when it is unset (trailing slash trimmed).
2. Derives `DATABASE_URL` from the `POSTGRES_*` variables, URL-encoding user and password and
   defaulting `POSTGRES_DB` to `opus_go` and the host to `postgres`. **Exits 1** when neither
   `DATABASE_URL` nor `POSTGRES_PASSWORD` is set. `GO_DATABASE_URL`, when set, overrides the
   result.
3. Exports `PORT` (default `27538`), `CLIENT_DIR` (default `/client`) and `UPLOAD_DIR`
   (default `/app/data/uploads`).
4. Runs `web-env.sh`, which rewrites the `OPUS_API_URL`, `OPUS_CLIENT_URL` and other `OPUS_*`
   placeholders inside the built JavaScript and CSS.
5. `exec`s `/usr/local/bin/opus-api`.

`GET /api/health` returns `{"status":"ok"}` and `GET /api/ready` pings the database, answering
`503` until it is reachable.

## Databases

The Go backend keeps its own schema on its own database (`opus_go`, suite int64 ids) separate from
the legacy TypeScript database (`opus`, CUID2). The legacy backend has been decommissioned, so a
fresh install only needs `opus_go`.

At startup the API checks for the database and creates it when missing, so an existing PostgreSQL
volume — where `deploy/db/init/01-opus-go.sql` never ran, because init scripts only fire on a fresh
volume — no longer blocks the boot. If the role lacks `CREATEDB`, creation is skipped and the
underlying connection error is logged instead. To create it by hand:

```sh
docker exec opus-db psql -U opus -c 'CREATE DATABASE opus_go'
```

Two failure modes to know:

- **`password authentication failed for user "opus"`** — `POSTGRES_PASSWORD` changed after the
  volume was initialised. Postgres applies that password only on first init; an existing volume
  keeps the original role password. Keep the env var stable, or fix it in place with
  `ALTER USER opus WITH PASSWORD '…'`.
- **`tls error: server refused TLS connection`** — benign. pgx defaults to `sslmode=prefer`, tries
  TLS, the stock Postgres image has none, and it falls back to plaintext.

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
  `/app/data/uploads` for uploaded assets, restarts unless stopped, and repeats the image's
  healthcheck.

Neither service publishes a host port. The `opus` container is reachable only on the compose
network, which is what you want behind Traefik and what you must change for a direct local run.

A commented-out `redis` service (Valkey) is included for the HA setup described below.

## Behind Traefik on la ruche

Route one hostname to the `opus` container's **port `27538`** — the port the image listens on.
Nothing else in the container should be exposed. The variable that matters for the
public hostname:

```sh
OPUS_CLIENT_URL=https://opus.facile.studio
```

`OPUS_API_URL` is then derived as `https://opus.facile.studio/api`, and CORS falls back to
`OPUS_CLIENT_URL`. Setting `CORS_ORIGINS` explicitly is still a good idea in production — leaving
both unset makes the API echo back any origin. Because the client and the API share a hostname,
cookies are issued `SameSite=Lax` without `Secure`.

**A `404 page not found` from the edge is a routing problem, not an application one.** If every path
— including `/api/health` — answers with a short `text/plain` body and `x-content-type-options:
nosniff`, Traefik has no router for the host: the domain is not bound to the running service, or it
still points at a container that the Go cutover replaced. Check this before debugging the binary.
In Dokploy, open the application (or the `opus` service of the Compose app), go to **Domains**, and
make sure the hostname is mapped to container port `27538`. Recreating a Dokploy application drops its
domain configuration, so re-add it after a rebuild.

A healthy container is not proof the front end is reachable: `GET /` must return the SPA's
`index.html`, and `/api/health` must return `200`. Verify both from outside the host, not just
`127.0.0.1` inside it.

## Scaling out

The default WebSocket broadcast adapter is in-memory, so more than one replica silently loses
real-time updates between them. Set `REDIS_URL` (or the sentinel or cluster variables) before
raising the replica count. Uploads have the same problem in reverse: assets are written to local
disk (`UPLOAD_DIR`), so either keep a single replica or back it with shared storage.

## Kubernetes

`charts/opus` is a self-contained chart: Deployment, Service, ServiceAccount, optional PVC,
optional HPA, and both an Ingress and a Gateway API HTTPRoute. `postgresql.enabled` defaults to
`true` and deploys `postgres:16-alpine` inside the release with an 8 Gi PVC; point it at an
external database for anything serious.

```sh
helm install opus ./charts/opus --set opus.env.OPUS_CLIENT_URL=https://opus.example.com
```

The chart's `opus.service.port` and `opus.service.targetPort` default to `27538`, matching the port
the image listens on; the Ingress and HTTPRoute backends use the same value.

One thing to check before installing:

- The default Postgres credentials in `values.yaml` are `opus_user` / `opus_password` and create the
  `opus` database. Point the application at `opus_go` and let startup create it, or use
  `postgresql.auth.existingSecret`.

The chart's liveness and readiness probes already target `/api/health`. `opus.extraEnv` takes
arbitrary environment variables for the application container.
