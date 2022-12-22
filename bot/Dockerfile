FROM golang:1.17-alpine

WORKDIR /app
COPY . .
RUN go build -o bin/bot bot/main.go bot/bot.go
