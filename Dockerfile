# syntax=docker/dockerfile:1

# --- Dev stage: hot reload with air (only built with --target dev) ---
FROM golang:1.27.2-alpine AS dev

RUN apk add --no-cache git libwebp-tools

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Source is bind-mounted at runtime (see docker-compose.dev.yml)
CMD ["air", "-c", ".air.toml"]


# --- Stage 1: Build the Go binary ---
FROM golang:1.27.2-alpine AS builder

WORKDIR /app

# Cache dependencies separately
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build binary with cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false -ldflags="-s -w" -o server ./main.go


# --- Stage 2: Runtime image (must stay last: it is the default target) ---
FROM alpine:3.20

WORKDIR /app

# Install WebP tools in the runtime image
RUN apk add --no-cache libwebp-tools

# Copy compiled Go binary
COPY --from=builder /app/server .

EXPOSE 8081

CMD ["./server"]
