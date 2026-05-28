# syntax=docker/dockerfile:1
FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/scheduler ./cmd/scheduler

FROM alpine:latest

WORKDIR /

RUN apk add --no-cache netcat-openbsd

COPY --from=builder /app/scheduler /scheduler
COPY --from=builder /app/configs/scheduler_config.yaml /scheduler_config.yaml
COPY docker/scripts/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

CMD ["/wait-for-it.sh", "postgres", "5432", "--", "/scheduler", "--config", "/scheduler_config.yaml"]
