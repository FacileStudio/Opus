# Build the transitional TypeScript API (apps/api-legacy). The Go backend
# proxies every /api route it does not own yet to this build until the port
# is complete.

FROM oven/bun:1 AS api-builder
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/typescript-config/package.json ./packages/typescript-config/
COPY packages/email/package.json ./packages/email/
COPY packages/libs/package.json ./packages/libs/
COPY packages/mcp/package.json ./packages/mcp/
COPY apps/api-legacy/package.json ./apps/api-legacy/
COPY apps/web/package.json ./apps/web/
COPY apps/site/package.json ./apps/site/

RUN pnpm install --frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY packages/email ./packages/email
COPY packages/libs ./packages/libs
COPY apps/api-legacy ./apps/api-legacy

RUN cd packages/email && pnpm run build
RUN cd apps/api-legacy && pnpm run build

FROM oven/bun:1 AS api-deps
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY --from=api-builder /app/package.json /app/pnpm-workspace.yaml /app/pnpm-lock.yaml ./
COPY --from=api-builder /app/packages/typescript-config/package.json ./packages/typescript-config/
COPY --from=api-builder /app/packages/email/package.json ./packages/email/
COPY --from=api-builder /app/packages/libs/package.json ./packages/libs/
COPY --from=api-builder /app/packages/mcp/package.json ./packages/mcp/
COPY --from=api-builder /app/apps/api-legacy/package.json ./apps/api-legacy/
COPY --from=api-builder /app/apps/web/package.json ./apps/web/
COPY --from=api-builder /app/apps/site/package.json ./apps/site/

# Do not put `--no-optional` back: @node-rs/argon2 ships its compiled binding
# as a per-platform optionalDependency, so the flag installs the JS wrapper
# with nothing to load and the API dies on boot.
RUN NODE_ENV=production pnpm install --prod --frozen-lockfile

# The legacy runtime stays at /app/apps/api so its uploads, migrations and
# relative paths match the pre-port deployment; only the source directory
# moved to apps/api-legacy. The node_modules symlink bridges pnpm's
# per-workspace dependency links for the relocated bundle.
RUN mkdir -p apps/api && ln -s ../api-legacy/node_modules apps/api/node_modules
COPY --from=api-builder /app/apps/api-legacy/dist ./apps/api/dist
COPY --from=api-builder /app/apps/api-legacy/drizzle ./apps/api/drizzle
COPY --from=api-builder /app/packages/email ./packages/email

FROM oven/bun:1 AS web-builder
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/typescript-config/package.json ./packages/typescript-config/
COPY packages/email/package.json ./packages/email/
COPY packages/libs/package.json ./packages/libs/
COPY packages/mcp/package.json ./packages/mcp/
COPY apps/api-legacy/package.json ./apps/api-legacy/
COPY apps/web/package.json ./apps/web/
COPY apps/site/package.json ./apps/site/

RUN pnpm install --frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY packages/libs ./packages/libs
COPY i18n ./i18n
COPY apps/web ./apps/web

RUN cd apps/web && NODE_ENV=production pnpm run build

# Build the new Go backend on the tronc chassis.
FROM golang:1.26-alpine AS go-builder

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o bin/api .

# Runtime: one container, one router. The Go binary serves the SPA and /api,
# proxying unported routes to the legacy API on 127.0.0.1:1337.

FROM oven/bun:1-slim AS runtime

RUN apt-get update && \
  apt-get install -y --no-install-recommends wget && \
  rm -rf /var/lib/apt/lists/* && \
  groupadd -g 1001 appuser && \
  useradd -u 1001 -g appuser -d /home/appuser -m appuser

WORKDIR /app

COPY --from=api-deps --chown=appuser:appuser /app /app
COPY --from=web-builder --chown=appuser:appuser /app/apps/web/dist /client
COPY --from=go-builder --chown=appuser:appuser /repo/apps/api/bin/api /usr/local/bin/opus-api
COPY --chown=appuser:appuser deploy/opus-entrypoint.sh /usr/local/bin/opus-entrypoint.sh
COPY --chown=appuser:appuser apps/web/env.sh /usr/local/bin/web-env.sh

RUN chmod +x /usr/local/bin/opus-entrypoint.sh /usr/local/bin/web-env.sh

ENV NODE_ENV=production

USER appuser
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:80/api/health || exit 1

CMD ["/usr/local/bin/opus-entrypoint.sh"]
