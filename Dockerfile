FROM golang:1.23.5

WORKDIR /app
RUN go install github.com/go-task/task/v3/cmd/task@latest
COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN go build -o ./bin/api cmd/api/main.go
RUN go build -o ./bin/cron cmd/cron/main.go

RUN chmod +x ./run.sh
RUN task install:migrator