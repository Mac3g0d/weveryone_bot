FROM golang:latest as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bot ./bot.go

FROM alpine:latest

COPY --from=0 /app/bot /bot

CMD ["./bot"]