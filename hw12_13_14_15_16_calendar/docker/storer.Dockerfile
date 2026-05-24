# syntax=docker/dockerfile:1
FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/storer ./cmd/storer

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/storer /storer
COPY --from=builder /app/configs/storer_config.yaml /storer_config.yaml

CMD ["/storer", "--config", "/storer_config.yaml"]
