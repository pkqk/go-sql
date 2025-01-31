FROM golang:1.19.0-alpine3.16

ENV CGO_ENABLED=0

RUN apk add --no-cache mysql-client postgresql-client

ENTRYPOINT [ "go", "test", "-v", "." ]
