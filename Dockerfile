# First stage: builder
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

# Install necessary certificates and build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /usr/src/app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy the rest of the application
COPY . .
RUN go build -v -o /run-app .

# Second stage: final image
FROM debian:bookworm-slim

COPY --from=builder /usr/src/app/components/*.txt /usr/src/app/components/

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application binary
COPY --from=builder /run-app /usr/local/bin/

# Copy migrate binary
COPY --from=builder /go/bin/migrate /usr/local/bin/
COPY --from=builder /usr/src/app/migrate.sh /app/migrate.sh
RUN chmod +x /app/migrate.sh

# Create migrations directory and copy migrations
RUN mkdir -p /app/migrations
COPY --from=builder /usr/src/app/internal/db/migrations/*.sql /app/migrations/

# Verify migrations are present (will fail build if migrations are missing)
RUN ls -la /app/migrations/

# Create empty .env file
RUN touch .env

# Set working directory
WORKDIR /app

CMD ["run-app"]
