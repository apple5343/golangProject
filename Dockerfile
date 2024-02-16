FROM golang:latest as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o cmd/main ./cmd

CMD ["./cmd/main"]