# Stage 1: Build
FROM golang:1.25 AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries in cmd directory
RUN go build -o bin/ ./cmd/...

# Stage 2: Runtime
FROM ubuntu:24.04

# Install curl for debugging
RUN apt-get update && \
    apt-get install -y curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy all compiled binaries from builder
COPY --from=builder /build/bin/* /app/

# Set octollm-server as default entrypoint
ENTRYPOINT ["/app/octollm-server"]

# Default command (can be overridden)
CMD []
