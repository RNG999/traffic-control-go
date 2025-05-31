# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o tcctl ./cmd/tcctl

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/tcctl /tcctl

# Expose any necessary ports (if needed for future web interface)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/tcctl"]

# Metadata labels
LABEL org.opencontainers.image.title="Traffic Control CLI Tool"
LABEL org.opencontainers.image.description="Human-readable Linux Traffic Control (TC) configuration tool"
LABEL org.opencontainers.image.vendor="rng999"
LABEL org.opencontainers.image.licenses="MIT"