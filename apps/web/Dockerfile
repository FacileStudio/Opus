FROM --platform=$BUILDPLATFORM oven/bun:1 AS builder

WORKDIR /app

COPY package.json bun.lockb ./

COPY packages ./packages
COPY apps/api ./apps/api
COPY apps/web ./apps/web

RUN bun install --no-frozen-lockfile

WORKDIR /app/apps/web
RUN bun run build

FROM nginx:alpine

COPY --from=builder /app/apps/web/dist /usr/share/nginx/html

COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf

COPY apps/web/env.sh /docker-entrypoint.d/env.sh
RUN chmod +x /docker-entrypoint.d/env.sh

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]