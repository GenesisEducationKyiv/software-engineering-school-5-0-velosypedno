FROM golang:1.23.5 AS builder
WORKDIR /app

COPY . .
WORKDIR /app/rmq-initializer
RUN go mod download 

RUN go build -o ./bin/rmq-initializer main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/rmq-initializer/bin/rmq-initializer ./bin/rmq-initializer

