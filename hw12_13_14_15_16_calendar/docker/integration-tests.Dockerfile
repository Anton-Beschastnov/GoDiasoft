# syntax=docker/dockerfile:1
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run the integration tests with a 5-minute timeout
CMD ["go", "test", "-v", "-timeout", "5m", "./tests/integration/..."]
