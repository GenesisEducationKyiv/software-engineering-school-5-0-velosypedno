FROM golang:1.23.5 AS builder
WORKDIR /app

COPY . .
WORKDIR /app/notifier
RUN go mod download 

RUN go build -o ./bin/notifier cmd/main.go

FROM debian:bookworm
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/

COPY --from=builder /app/notifier/bin/notifier ./bin/notifier

COPY notifier/internal/templates ./internal/templates