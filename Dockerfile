# syntax=docker/dockerfile:1

FROM golang:1.22.3-alpine AS build

RUN go version
ENV GOPATH=/

COPY ["cmd",                    "/app/server/advertd/cmd"]
COPY ["internal",               "/app/server/advertd/internal"]
COPY ["pkg",                    "/app/server/advertd/pkg"]
COPY ["settings.docker.json",   "/app/server/advertd/settings.json"]

# Add CGO compiler
RUN apk add build-base





# build go app
WORKDIR /app/server/advertd/cmd
RUN go mod tidy -e
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o advertd ./advertd.go

# Final stage
FROM alpine:3.20.2 AS run

WORKDIR /
COPY --from=build /app/server/advertd/cmd/advertd     /app/server/advertd/cmd/advertd
COPY --from=build /app/server/advertd/settings.json    /app/server/advertd/cmd/settings.json