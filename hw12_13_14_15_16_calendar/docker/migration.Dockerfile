# syntax=docker/dockerfile:1
FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache netcat-openbsd

COPY go.mod go.sum ./
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.20.0

COPY migrations ./migrations
COPY docker/scripts/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

CMD ["/wait-for-it.sh", "postgres", "5432", "--", "goose", "-dir", "./migrations", "postgres", "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable", "up"]
