FROM --platform=$BUILDPLATFORM oven/bun:1 AS base
WORKDIR /app

COPY package.json .
COPY bun.lockb .
COPY apps/api/package.json ./apps/api/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json

RUN bun install --no-frozen-lockfile

COPY packages/typescript-config ./packages/typescript-config
COPY apps/api ./apps/api

FROM oven/bun:1-alpine AS runtime
WORKDIR /app

COPY --from=base /app ./

RUN mkdir -p /data && chown -R nobody:nobody /data

ENV NODE_ENV=production
WORKDIR /app/apps/api

USER nobody

EXPOSE 1337

CMD ["bun", "run", "src/index.ts"]