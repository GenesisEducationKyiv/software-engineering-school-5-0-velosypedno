FROM golang:1.23.5 AS builder
WORKDIR /app

RUN go install github.com/go-task/task/v3/cmd/task@latest
COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN go build -o ./bin/api cmd/api/main.go
RUN go build -o ./bin/cron cmd/cron/main.go
RUN task install:migrator

FROM debian:bookworm
WORKDIR /app

COPY --from=builder /app/Taskfile.yml ./Taskfile.yml
COPY --from=builder /app/Taskfile.docker.yml ./Taskfile.docker.yml

COPY --from=builder /app/bin/api ./bin/api
COPY --from=builder /app/bin/cron ./bin/cron
COPY --from=builder /go/bin/task /usr/local/bin/task
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY --from=builder /app/db/migrations ./db/migrations

