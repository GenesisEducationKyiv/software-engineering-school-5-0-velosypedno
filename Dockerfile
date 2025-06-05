FROM golang:1.23.5

WORKDIR /app
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && migrate -help
COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN go build -o ./bin/api cmd/api/main.go
RUN go build -o ./bin/cron cmd/cron/main.go

RUN chmod +x ./run.sh
RUN chmod +x ./migrate.sh