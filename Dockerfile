# Quiver - An SSH TUI Application
# Copyright (C) 2026  penaz
# SPDX-License-Identifier: GPL-3.0-or-later

# ─── Stage 1: Build ──────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /bin/quiver .

# ─── Stage 2: Runtime ────────────────────────────────────────────────
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S quiver \
    && adduser -S -G quiver -h /home/quiver -s /sbin/nologin quiver

# Persistent data volume mount point
RUN mkdir -p /data && chown quiver:quiver /data
VOLUME /data

COPY --from=builder /bin/quiver /usr/local/bin/quiver

USER quiver

# Default environment — all overridable at runtime
ENV QUIVER_HOST=0.0.0.0
ENV QUIVER_PORT=2222
ENV QUIVER_DATA_DIR=/data

EXPOSE 2222

ENTRYPOINT ["quiver"]
