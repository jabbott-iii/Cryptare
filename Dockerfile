# Builder: the same Go release as the toolchain line in go.mod (SEC-018), pinned by
# digest (SEC-013). Change the tag and digest together with go.mod.
FROM golang:1.26.8-alpine@sha256:8ac98ca534ac3f51e1f420a1dd2c15e74c75cfa0f23f3ad27eb5d7236c349a0c AS builder

WORKDIR /src

# Install build deps for CGO sqlite3 driver
RUN apk add --no-cache build-base

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build. VERSION is what --version reports (default "dev").
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-X main.version=${VERSION}" -o /out/cryptare .

# Runtime, pinned by digest (SEC-013). No packages are added: go-sqlite3 compiles
# SQLite into the binary, and Cryptare makes no network connections.
FROM alpine:3.22@sha256:5291449c3df73caf6ed85e649dec1b9e818b39a5d8c871e97afc13e9cd5e8fa8

# An unprivileged user that owns the data folder (SEC-013). A new named volume
# takes this ownership; for a bind mount, run with --user "$(id -u):$(id -g)".
RUN addgroup -S -g 10001 cryptare \
 && adduser -S -D -H -u 10001 -G cryptare cryptare \
 && mkdir -p /app/data \
 && chown cryptare:cryptare /app/data \
 && chmod 0700 /app/data

WORKDIR /app
COPY --from=builder /out/cryptare /usr/local/bin/cryptare

# Persist sqlite database file (cryptare.db)
VOLUME ["/app/data"]

ENV CRYPTARE_DB_PATH=/app/data/cryptare.db

USER 10001:10001

# This app is an interactive TUI/CLI
ENTRYPOINT ["cryptare"]
