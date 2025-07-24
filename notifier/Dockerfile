FROM golang:1.23.5 AS builder
WORKDIR /app

COPY . .
WORKDIR /app/notifier
RUN go mod download 

RUN go build -o ./bin/notifier cmd/main.go

FROM debian:bookworm
WORKDIR /app

COPY --from=builder /app/notifier/bin/notifier ./bin/notifier

