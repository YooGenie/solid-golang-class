## Build Stage
FROM golang:1.17-alpine3.15 AS builder 

RUN apk update \
    && apk add build-base

WORKDIR /app

COPY go.mod go.sum main.go scripts/setup.sh ./
COPY pkg pkg
COPY cmd cmd
COPY vendor vendor
RUN mkdir -p configs
RUN chmod +x setup.sh \
    && ./setup.sh

## Deploy Stage
FROM alpine:3.15.0

WORKDIR /app
RUN adduser -D edpuser && chown -R edpuser /app
COPY --chown=root:root --from=builder /app/bin/event-data-pipeline ./
COPY --chown=root:root --from=builder /app/configs ./configs
EXPOSE 8078
USER edpuser

ENTRYPOINT [ "./event-data-pipeline" ]