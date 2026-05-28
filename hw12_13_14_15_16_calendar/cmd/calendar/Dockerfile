# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/calendar ./cmd/calendar

# Final stage
FROM alpine:latest

COPY --from=builder /bin/calendar /bin/calendar
COPY --from=builder /app/configs/ /configs/

EXPOSE 8080

ENTRYPOINT ["/bin/calendar"]
