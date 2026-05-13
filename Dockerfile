FROM oven/bun:1 AS api-builder
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/typescript-config/package.json ./packages/typescript-config/
COPY packages/email/package.json ./packages/email/
COPY packages/libs/package.json ./packages/libs/
COPY packages/mcp/package.json ./packages/mcp/
COPY apps/api/package.json ./apps/api/
COPY apps/web/package.json ./apps/web/
COPY apps/site/package.json ./apps/site/

RUN pnpm install --frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY packages/email ./packages/email
COPY packages/libs ./packages/libs
COPY apps/api ./apps/api

RUN cd packages/email && pnpm run build
RUN cd apps/api && pnpm run build

FROM oven/bun:1 AS api-deps
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY --from=api-builder /app/package.json /app/pnpm-workspace.yaml /app/pnpm-lock.yaml ./
COPY --from=api-builder /app/packages/typescript-config/package.json ./packages/typescript-config/
COPY --from=api-builder /app/packages/email/package.json ./packages/email/
COPY --from=api-builder /app/packages/libs/package.json ./packages/libs/
COPY --from=api-builder /app/packages/mcp/package.json ./packages/mcp/
COPY --from=api-builder /app/apps/api/package.json ./apps/api/
COPY --from=api-builder /app/apps/web/package.json ./apps/web/
COPY --from=api-builder /app/apps/site/package.json ./apps/site/

RUN sed -i 's/"prepare": "husky"/"prepare": ""/g' package.json && \
  HUSKY=0 NODE_ENV=production pnpm install --prod --frozen-lockfile --no-optional

COPY --from=api-builder /app/apps/api/dist ./apps/api/dist
COPY --from=api-builder /app/apps/api/drizzle ./apps/api/drizzle
COPY --from=api-builder /app/packages/email ./packages/email

FROM oven/bun:1 AS web-builder
WORKDIR /app

RUN bun install -g pnpm@10.32.1

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY packages/typescript-config/package.json ./packages/typescript-config/
COPY packages/email/package.json ./packages/email/
COPY packages/libs/package.json ./packages/libs/
COPY packages/mcp/package.json ./packages/mcp/
COPY apps/api/package.json ./apps/api/
COPY apps/web/package.json ./apps/web/
COPY apps/site/package.json ./apps/site/

RUN pnpm install --frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY packages/libs ./packages/libs
COPY i18n ./i18n
COPY apps/web ./apps/web

RUN cd apps/web && NODE_ENV=production pnpm run build

FROM oven/bun:1-slim AS runtime

RUN apt-get update && \
  apt-get install -y --no-install-recommends nginx wget && \
  rm -rf /var/lib/apt/lists/* && \
  groupadd -g 1001 appuser && \
  useradd -u 1001 -g appuser -d /home/appuser -m appuser && \
  mkdir -p /app /app/apps/api/data /app/apps/api/data/uploads /var/cache/nginx /run/nginx /var/log/nginx /var/lib/nginx/body /var/lib/nginx/proxy /var/lib/nginx/fastcgi /var/lib/nginx/uwsgi /var/lib/nginx/scgi /usr/share/nginx/html /docker-entrypoint.d && \
  chown -R appuser:appuser /app /var/cache/nginx /run/nginx /var/log/nginx /var/lib/nginx /usr/share/nginx/html /docker-entrypoint.d && \
  touch /run/nginx.pid && \
  chown appuser:appuser /run/nginx.pid && \
  rm -f /etc/nginx/sites-enabled/default && \
  printf 'worker_processes auto;\nerror_log /var/log/nginx/error.log warn;\npid /run/nginx.pid;\nevents { worker_connections 1024; }\nhttp {\n  include /etc/nginx/mime.types;\n  default_type application/octet-stream;\n  include /etc/nginx/conf.d/*.conf;\n}\n' > /etc/nginx/nginx.conf

WORKDIR /app

COPY --from=api-deps --chown=appuser:appuser /app /app
COPY --from=web-builder --chown=appuser:appuser /app/apps/web/dist /usr/share/nginx/html
COPY --chown=appuser:appuser apps/web/nginx.opus.conf /etc/nginx/conf.d/default.conf
RUN chown -R appuser:appuser /etc/nginx/conf.d
COPY --chown=appuser:appuser apps/web/env.sh /docker-entrypoint.d/env.sh
COPY --chown=appuser:appuser deploy/opus-entrypoint.sh /usr/local/bin/opus-entrypoint.sh

RUN chmod +x /docker-entrypoint.d/env.sh /usr/local/bin/opus-entrypoint.sh

ENV NODE_ENV=production

USER appuser
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:80/api/health || exit 1

CMD ["/usr/local/bin/opus-entrypoint.sh"]
