# syntax=docker/dockerfile:1
FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run the integration tests
CMD ["go", "test", "-v", "./tests/integration/..."]
