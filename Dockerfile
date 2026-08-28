# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS frontend
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@10.15.0 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

FROM golang:1.25-alpine AS backend
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY backend/ ./
COPY --from=frontend /src/frontend/dist/ ./cmd/server/static/
ARG VERSION=dev
ARG COMMIT=unknown
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /out/dashlit ./cmd/server

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 dashlit \
    && adduser -S -D -H -u 10001 -G dashlit dashlit \
    && mkdir -p /data \
    && chown -R dashlit:dashlit /data
WORKDIR /app
COPY --from=backend /out/dashlit /usr/local/bin/dashlit
ENV ADDR=:8080 DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
USER dashlit
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -T 4 -O /dev/null http://127.0.0.1:8080/api/auth/config || exit 1
ENTRYPOINT ["/usr/local/bin/dashlit"]
