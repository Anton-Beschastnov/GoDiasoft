# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/scheduler ./cmd/scheduler

# Final stage
FROM alpine:latest

COPY --from=builder /bin/scheduler /bin/scheduler
COPY --from=builder /app/configs/ /configs/

ENTRYPOINT ["/bin/scheduler"]
