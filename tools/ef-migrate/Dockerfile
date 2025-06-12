# Multi-stage Dockerfile for EF Migration Tool
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Set build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies (sorted alphabetically)
RUN apk add --no-cache ca-certificates git tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code (using .dockerignore to exclude sensitive files)
COPY . .

# Build the EF migration tool
WORKDIR /app/tools/ef-migrate
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -w -s" \
    -a -installsuffix cgo \
    -o ef-migrate .

# Stage 2: Create the final lightweight image
FROM alpine:3.19

# Install runtime dependencies (sorted alphabetically) and create non-root user
RUN apk add --no-cache \
    ca-certificates \
    postgresql-client \
    sqlite \
    tzdata \
    && rm -rf /var/cache/apk/* \
    && addgroup -g 1001 -S efmigrate \
    && adduser -u 1001 -S efmigrate -G efmigrate

# Set working directory
WORKDIR /app

# Copy the binary from builder stage and set up application structure
COPY --from=builder /app/tools/ef-migrate/ef-migrate /usr/local/bin/ef-migrate

# Create directories and set permissions in single RUN
RUN mkdir -p /app/migrations \
    && chown -R efmigrate:efmigrate /app \
    && chmod +x /usr/local/bin/ef-migrate

# Switch to non-root user
USER efmigrate

# Set environment variables
ENV MIGRATIONS_DIR=/app/migrations
ENV PATH="/usr/local/bin:$PATH"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ef-migrate version || exit 1

# Default command
ENTRYPOINT ["ef-migrate"]
CMD ["help"]

# Metadata labels
LABEL maintainer="Lambok Tulus Simamora <your-email@example.com>"
LABEL description="GRA Framework EF Migration Tool - Entity Framework-like database migration utility"
LABEL version="${VERSION}"
LABEL org.opencontainers.image.title="GRA EF Migration Tool"
LABEL org.opencontainers.image.description="Entity Framework-like database migration utility for the GRA framework"
LABEL org.opencontainers.image.url="https://github.com/lamboktulussimamora/gra"
LABEL org.opencontainers.image.source="https://github.com/lamboktulussimamora/gra"
LABEL org.opencontainers.image.vendor="GRA Framework"
LABEL org.opencontainers.image.licenses="MIT"
