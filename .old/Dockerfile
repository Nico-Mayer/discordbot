FROM golang:1.24.1-alpine AS builder

# Add edge repositories and install dependencies
RUN apk update && \
    apk add --no-cache \
    --repository=https://dl-cdn.alpinelinux.org/alpine/edge/main \
    --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    ffmpeg \
    yt-dlp

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o out .

# Use multi-stage build for smaller final image
FROM alpine:edge

# Install runtime dependencies from edge
RUN apk add --no-cache \
    ffmpeg \
    yt-dlp

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /app/out .

# Run as non-root user for security
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser && \
    chown -R appuser:appgroup /app

USER appuser

ENTRYPOINT ["/app/out"]
