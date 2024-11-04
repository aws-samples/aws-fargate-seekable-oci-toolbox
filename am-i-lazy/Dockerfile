# syntax=docker/dockerfile:1.5
## Build
FROM public.ecr.aws/docker/library/golang:1.23-alpine AS build

WORKDIR /app

COPY app/go.mod ./
COPY app/go.sum ./
ENV GOPROXY=direct
RUN apk add git && \
  go mod download

COPY app/*.go ./

RUN go build -o /amilazy

## Deploy
FROM public.ecr.aws/docker/library/alpine:3.20

WORKDIR /

COPY --from=build /amilazy /amilazy
USER nobody:nobody

ENTRYPOINT ["/amilazy"]