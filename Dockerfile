# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/trakkr ./cmd/trakkr && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/trakkr-server ./cmd/trakkr-server

FROM alpine:3.22

RUN addgroup -S trakkr && adduser -S -G trakkr trakkr

WORKDIR /app
COPY --from=build /out/trakkr /usr/local/bin/trakkr
COPY --from=build /out/trakkr-server /usr/local/bin/trakkr-server
COPY migrations ./migrations
COPY web ./web

USER trakkr
EXPOSE 8080

ENV TRAKKR_ADDR=:8080
ENTRYPOINT ["/usr/local/bin/trakkr-server"]
