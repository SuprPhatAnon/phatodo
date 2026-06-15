# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/phatodo ./cmd/phatodo && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/ptd ./cmd/ptd && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/phatodo-server ./cmd/phatodo-server

FROM alpine:3.22

RUN addgroup -S phatodo && adduser -S -G phatodo phatodo

WORKDIR /app
COPY --from=build /out/phatodo /usr/local/bin/phatodo
COPY --from=build /out/ptd /usr/local/bin/ptd
COPY --from=build /out/phatodo-server /usr/local/bin/phatodo-server
COPY migrations ./migrations
COPY web ./web

USER phatodo
EXPOSE 8080

ENV PHATODO_ADDR=:8080
ENTRYPOINT ["/usr/local/bin/phatodo-server"]
