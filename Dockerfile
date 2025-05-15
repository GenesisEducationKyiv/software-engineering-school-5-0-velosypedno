FROM golang:1.23.5

WORKDIR /app
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && migrate -help
COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN go build -o ./bin/main cmd/weather/main.go

RUN chmod +x ./run.sh
ENTRYPOINT [ "./run.sh" ]