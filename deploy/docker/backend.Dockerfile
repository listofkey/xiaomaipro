# syntax=docker/dockerfile:1.7

FROM golang:1.25.5-alpine AS builder

WORKDIR /src/server

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server ./

ARG SERVICE_PATH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/service ${SERVICE_PATH}

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/service /app/service
COPY server/deploy/config /app/config

ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/service"]
