FROM node:22-alpine AS builder

RUN npm install -g bun

RUN apk add --no-cache \
    python3 \
    make \
    g++

WORKDIR /app

COPY package.json bun.lockb ./

COPY packages ./packages
COPY apps/api ./apps/api
COPY apps/web ./apps/web

RUN bun install

WORKDIR /app/apps/web
RUN bun run build

FROM nginx:alpine

COPY --from=builder /app/apps/web/dist /usr/share/nginx/html

COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf

COPY apps/web/env.sh /docker-entrypoint.d/env.sh
RUN chmod +x /docker-entrypoint.d/env.sh

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]