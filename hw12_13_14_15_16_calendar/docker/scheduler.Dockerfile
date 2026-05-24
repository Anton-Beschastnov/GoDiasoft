# syntax=docker/dockerfile:1
FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/scheduler ./cmd/scheduler

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/scheduler /scheduler
COPY --from=builder /app/configs/scheduler_config.yaml /scheduler_config.yaml

CMD ["/scheduler", "--config", "/scheduler_config.yaml"]
