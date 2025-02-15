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

ENV NODE_ENV=production
WORKDIR /app/apps/api

EXPOSE 1337

CMD ["bun", "run", "src/index.ts"]