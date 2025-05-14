FROM golang:1.23.5
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download 
COPY . . 
RUN go build -o ./bin/main main.go
RUN chmod +x ./run.sh
ENTRYPOINT [ "./run.sh" ]
EXPOSE 8080