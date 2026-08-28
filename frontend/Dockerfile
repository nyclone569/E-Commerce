# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS dependencies
WORKDIR /app
RUN npm install --global pnpm@11.23.0
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

FROM node:24-alpine AS build
WORKDIR /app
RUN npm install --global pnpm@11.23.0
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
ARG APP_VERSION=dev
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

FROM node:24-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0
RUN addgroup -S aurora && adduser -S -G aurora aurora
COPY --from=build --chown=aurora:aurora /app/public ./public
COPY --from=build --chown=aurora:aurora /app/.next/standalone ./
COPY --from=build --chown=aurora:aurora /app/.next/static ./.next/static
USER aurora
EXPOSE 3000
CMD ["node", "server.js"]
