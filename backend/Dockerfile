# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget && addgroup -S aurora && adduser -S -G aurora aurora
COPY --from=build /out/api /usr/local/bin/api
USER aurora
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
