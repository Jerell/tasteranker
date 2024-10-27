ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /run-app /usr/local/bin/
# Copy the migrate binary
COPY --from=builder /go/bin/migrate /usr/local/bin/
# Copy migrations to a specific folder
COPY --from=builder /usr/src/app/internal/db/migrations /app/migrations
RUN touch .env
CMD ["run-app"]
