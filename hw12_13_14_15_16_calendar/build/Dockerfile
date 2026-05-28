# Build stage
FROM golang:1.22 as build

ARG APP_NAME
ENV BIN_FILE /opt/calendar/${APP_NAME}
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Cache modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Build the Go binary
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/${APP_NAME}/*.go

# Final stage
FROM alpine:3.15

ARG APP_NAME
LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE=${APP_NAME}
LABEL MAINTAINERS="student@otus.ru"

RUN apk add --no-cache gettext

ENV BIN_FILE "/opt/calendar/${APP_NAME}"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_TEMPLATE /etc/calendar/config.yaml.template
ENV CONFIG_FILE /etc/calendar/config.yaml

COPY build/entrypoint.sh /opt/entrypoint.sh
RUN chmod +x /opt/entrypoint.sh

# We will mount migrations for the calendar service, but let's copy them anyway
COPY migrations /opt/calendar/migrations
WORKDIR /opt/calendar

ENTRYPOINT ["/opt/entrypoint.sh"]
CMD ${BIN_FILE} -config ${CONFIG_FILE}