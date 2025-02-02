FROM --platform=$BUILDPLATFORM node:22-slim AS base
WORKDIR /app

COPY package.json .
COPY bun.lockb .
COPY apps/api/package.json ./apps/api/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json

RUN npm i -g bun

FROM base AS all-deps
RUN bun install

FROM base AS prod-deps
RUN bun install

FROM all-deps AS build
ENV NODE_ENV=production

COPY ./apps/api ./apps/api

FROM oven/bun:alpine AS runtime
WORKDIR /app

COPY --from=prod-deps /app/node_modules ./node_modules

COPY --from=build /app/apps/api ./apps/api

ENV NODE_ENV=production

WORKDIR /app/apps/api

EXPOSE 1337

CMD ["bun", "run", "src/index.ts"]
