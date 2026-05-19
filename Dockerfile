# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm AS gobuild
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY back/go.mod back/go.sum ./
RUN go mod download
COPY back/ ./
ENV CGO_ENABLED=0
RUN GOOS=linux go build -o /server ./cmd/server

FROM node:22-bookworm AS frontbuild
WORKDIR /app
COPY front/package.json front/pnpm-lock.yaml front/pnpm-workspace.yaml ./
RUN corepack enable pnpm && pnpm install --frozen-lockfile
COPY front/ ./
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

FROM node:22-bookworm-slim AS runner
WORKDIR /app/front
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=frontbuild /app/node_modules ./node_modules
COPY --from=frontbuild /app/package.json ./package.json
COPY --from=frontbuild /app/.next ./.next
COPY --from=frontbuild /app/public ./public
COPY --from=gobuild /server /app/server
WORKDIR /app
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh
ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3000
EXPOSE 3000
ENTRYPOINT ["/app/docker-entrypoint.sh"]
