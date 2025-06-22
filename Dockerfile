FROM golang:1.23.5 AS builder
WORKDIR /app

RUN go install github.com/go-task/task/v3/cmd/task@latest
COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN go build -o ./bin/api cmd/api/main.go
RUN task install:migrator

FROM debian:bookworm
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/

COPY --from=builder /app/Taskfile.yml ./Taskfile.yml
COPY --from=builder /app/Taskfile.docker.yml ./Taskfile.docker.yml

COPY --from=builder /app/bin/api ./bin/api
COPY --from=builder /go/bin/task /usr/local/bin/task
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY db/migrations ./db/migrations
COPY internal/templates ./internal/templates

