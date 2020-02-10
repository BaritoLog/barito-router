FROM golang:1.13-alpine AS build

RUN apk add build-base git && \
  adduser -Dh /app app
USER app

WORKDIR /app
COPY --chown=app:app go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p bin && \
  go build -o bin/ -ldflags "-linkmode external -extldflags -static" ./...

FROM scratch

COPY --from=build /app/bin/barito-router /bin/barito-router
COPY --from=build /etc/passwd /etc/group /etc/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER app
ENTRYPOINT [ "/bin/barito-router" ]
