# syntax=docker/dockerfile:1.5
## Build
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY app/go.mod ./
COPY app/go.sum ./
ENV GOPROXY=direct
RUN apk add git && \
  go mod download

COPY app/*.go ./

RUN go build -o /amilazy

## Deploy
FROM alpine:3.16

WORKDIR /

COPY --from=build /amilazy /amilazy
USER nobody:nobody

ENTRYPOINT ["/amilazy"]