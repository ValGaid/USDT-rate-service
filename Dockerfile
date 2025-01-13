# Этап сборки
FROM golang:1.23.2-alpine3.20 as builder

WORKDIR /usdt

COPY go.mod go.sum ./
RUN go mod download


COPY . .


ARG VERSION
RUN go build -v -ldflags="-X 'main.version=$VERSION'" -o /tmp/usdt ./cmd/

FROM alpine:3.20


RUN apk add --no-cache ca-certificates tzdata
RUN update-ca-certificates


COPY --from=builder /tmp/usdt /usr/bin/usdt

RUN chmod +x /usr/bin/usdt


RUN adduser -D -u 1000 usdt

USER usdt

COPY /migrate/20250106145630_create_usdt_table.sql /usdt/20250106145630_create_usdt_table.sql

COPY .env /usdt/.env

WORKDIR /usdt

ENTRYPOINT ["usdt"]
