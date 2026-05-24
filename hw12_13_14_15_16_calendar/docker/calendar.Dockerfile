# syntax=docker/dockerfile:1
FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/calendar ./cmd/calendar

FROM alpine:latest

WORKDIR /

# Устанавливаем netcat, необходимый для скрипта ожидания
RUN apk add --no-cache netcat-openbsd

COPY --from=builder /app/calendar /calendar
COPY --from=builder /app/configs/config.yaml /config.yaml
COPY --from=builder /app/migrations /migrations
COPY docker/scripts/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

EXPOSE 8888

CMD ["/wait-for-it.sh", "postgres", "5432", "--", "/calendar", "--config", "/config.yaml"]
