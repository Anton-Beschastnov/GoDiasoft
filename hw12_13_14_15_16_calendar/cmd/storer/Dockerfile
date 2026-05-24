# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/storer ./cmd/storer

# Final stage
FROM alpine:latest

COPY --from=builder /bin/storer /bin/storer
COPY --from=builder /app/configs/ /configs/

ENTRYPOINT ["/bin/storer"]
