FROM golang:1.23.5 AS builder
WORKDIR /app

RUN go install github.com/go-task/task/v3/cmd/task@latest
COPY . .
WORKDIR /app/sub
RUN go mod download 

RUN go build -o ./bin/sub cmd/main.go
RUN task install:migrator

FROM debian:bookworm
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/

COPY --from=builder /app/sub/taskfile.yml ./taskfile.yml

COPY --from=builder /app/sub/bin/sub ./bin/sub
COPY --from=builder /go/bin/task /usr/local/bin/task
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY sub/db/migrations ./db/migrations
COPY sub/internal/templates ./internal/templates

