FROM golang:1.19-alpine AS builder

WORKDIR /build

# Install timezone data
RUN apk update && \
    apk add ca-certificates && \
    apk add tzdata

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build \
    -o bverfgbot

FROM alpine:3.14

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/er
COPY --from=builder /build/bverfgbot /usr/bin/bverfgbot

ENTRYPOINT [ "bverfgbot" ]