# Stage 1: Build the React SPA
FROM oven/bun:1 AS web-builder
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/typescript-config/package.json ./packages/typescript-config/
COPY packages/email/package.json ./packages/email/
COPY packages/libs/package.json ./packages/libs/
COPY packages/mcp/package.json ./packages/mcp/
COPY apps/web/package.json ./apps/web/
COPY apps/site/package.json ./apps/site/

RUN pnpm install --frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY packages/libs ./packages/libs
COPY i18n ./i18n
COPY apps/web ./apps/web

RUN cd apps/web && NODE_ENV=production pnpm run build

# Stage 2: Build the Go backend on the tronc chassis
FROM golang:1.26-alpine AS go-builder

RUN apk add --no-cache git ca-certificates

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
COPY package.json /repo/package.json
COPY scripts/version.sh /repo/scripts/version.sh
RUN sh /repo/scripts/version.sh /repo/package.json > /repo/VERSION

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -X main.version=$(cat /repo/VERSION)" -o bin/api .

# Stage 3: Runtime container (single Go binary serving API and SPA)
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 1001 appuser && \
    adduser -u 1001 -G appuser -h /home/appuser -D appuser && \
    mkdir -p /app/data/uploads && \
    chown -R appuser:appuser /app

WORKDIR /app

COPY --from=web-builder --chown=appuser:appuser /app/apps/web/dist /client
COPY --from=go-builder --chown=appuser:appuser /repo/apps/api/bin/api /usr/local/bin/opus-api
COPY --chown=appuser:appuser deploy/opus-entrypoint.sh /usr/local/bin/opus-entrypoint.sh
COPY --chown=appuser:appuser apps/web/env.sh /usr/local/bin/web-env.sh

RUN chmod +x /usr/local/bin/opus-entrypoint.sh /usr/local/bin/web-env.sh

ENV NODE_ENV=production
ENV PORT=80
ENV CLIENT_DIR=/client

USER appuser
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD ["/usr/local/bin/opus-api", "healthcheck"]

CMD ["/usr/local/bin/opus-entrypoint.sh"]
