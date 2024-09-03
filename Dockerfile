# syntax=docker/dockerfile:1

FROM golang:1.22.3-alpine AS build

RUN go version
ENV GOPATH=/

COPY ["cmd",                   "/app/server/producer/cmd"]
COPY ["internal",              "/app/server/producer/internal"]
COPY ["pkg",                   "/app/server/producer/pkg"]
COPY ["settings.docker.json",  "/app/server/producer/settings.json"]

# Add CGO compiler
RUN apk add build-base




# build go app
WORKDIR /app/server/producer/cmd
RUN go mod tidy -e
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o producer ./producer.go

# Final stage
#FROM debian:buster AS run
FROM alpine:3.20.2 AS run

WORKDIR /
COPY --from=build /app/server/producer/cmd/producer     /app/server/producer/cmd/producer
COPY --from=build /app/server/producer/settings.json    /app/server/producer/cmd/settings.json