# syntax=docker/dockerfile:1
FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/calendar ./cmd/calendar

FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Устанавливаем netcat, необходимый для скрипта ожидания
RUN apk add --no-cache netcat-openbsd

# Копируем бинарник и конфигурацию в рабочую директорию
COPY --from=builder /app/calendar .
COPY --from=builder /app/configs/config.yaml .
# Копируем остальные ресурсы
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/swagger ./swagger
COPY docker/scripts/wait-for-it.sh .
RUN chmod +x ./wait-for-it.sh

EXPOSE 8888

# ИСПРАВЛЕНО: Запускаем из рабочей директории с правильными аргументами для wait-for-it
CMD ["./wait-for-it.sh", "postgres", "5432", "--", "./calendar", "--config", "./config.yaml"]
