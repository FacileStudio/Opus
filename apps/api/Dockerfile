FROM --platform=$BUILDPLATFORM node:22-slim AS base
WORKDIR /app

COPY package.json .
COPY bun.lockb .
COPY apps/api/package.json ./apps/api/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json

RUN npm i -g bun

RUN bun install --no-frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY apps/api ./apps/api

FROM oven/bun:alpine AS runtime
WORKDIR /app

COPY --from=base /app ./

RUN mkdir -p /app/apps/api/data && \
    chown -R nobody:nobody /app/apps/api/data

ENV NODE_ENV=production
WORKDIR /app/apps/api

USER nobody

EXPOSE 1337

CMD ["bun", "run", "src/index.ts"]