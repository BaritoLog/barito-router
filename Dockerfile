FROM golang:1.20-buster AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p bin && \
  go build -o bin/ ./...

FROM ubuntu:20.04

RUN useradd -m -U -d /app app
RUN apt update && apt install -y --no-install-recommends \
      ca-certificates && \
    apt-get clean && \
    rm -rf /tmp/* /var/tmp/* /var/lib/apt/lists/*
USER app

COPY --from=build /app/bin/barito-router /usr/bin/barito-router

ENTRYPOINT [ "/usr/bin/barito-router" ]
