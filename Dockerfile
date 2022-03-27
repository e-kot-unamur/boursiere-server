# syntax=docker/dockerfile:1

# https://docs.docker.com/language/golang/build-images/#multi-stage-builds

##
## Build
##
FROM golang:1.17-bullseye AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -ldflags '-s -w' -o /boursiere

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /boursiere ./
COPY sql/ ./sql/

ENV GIN_MODE=release
ENV PORT=80

EXPOSE 80
ENTRYPOINT ["/boursiere"]
