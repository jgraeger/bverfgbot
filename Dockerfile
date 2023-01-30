FROM golang:1.19-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build \
    -o bverfgbot

FROM alpine:3.14

COPY --from=builder /build/bverfgbot /usr/bin/bverfgbot

ENTRYPOINT [ "bverfgbot" ]