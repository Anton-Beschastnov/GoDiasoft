# syntax=docker/dockerfile:1
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.20.0

COPY migrations ./migrations

CMD goose -dir ./migrations postgres "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable" up
