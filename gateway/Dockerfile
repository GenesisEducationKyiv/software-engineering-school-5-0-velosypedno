FROM golang:1.23.5 AS builder
WORKDIR /app

COPY . .
WORKDIR /app/gateway
RUN go mod download 

RUN go build -o ./bin/gateway cmd/main.go

FROM debian:bookworm
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/

COPY --from=builder /app/gateway/bin/gateway ./bin/gateway

