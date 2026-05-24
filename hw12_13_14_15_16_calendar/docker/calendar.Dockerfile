# syntax=docker/dockerfile:1
FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/calendar ./cmd/calendar

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/calendar /calendar
COPY --from=builder /app/configs/config.yaml /config.yaml

EXPOSE 8888

CMD ["/calendar", "--config", "/config.yaml"]
