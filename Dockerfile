FROM golang:latest AS builder

RUN mkdir /app
COPY go.mod /app/
COPY go.sum /app/
WORKDIR /app
RUN go mod download
COPY ./ /app
RUN CGO_ENABLED=0 go build -o crontab-go
RUN chmod +x crontab-go


FROM debian:bookworm-slim AS slim

COPY --from=builder /app/crontab-go /bin/crontab-go

ENV LOG_TIMESTAMP_FORMAT="2006-01-02T15:04:05.000Z"
ENV LOG_FORMAT="ansi"
ENV LOG_FILE=
ENV LOG_STDOUT=true
ENV LOG_LEVEL="debug"

ENV SHELL="bash"
ENV SHELL_ARGS="-c"

ENV WEBSERVER_ADDRESS=
ENV WEBSERVER_PORT=
ENV WEBSERVER_USERNAME=
ENV WEBSERVER_PASSWORD=


ENTRYPOINT ["/bin/crontab-go" ]
CMD ["-c","/config.yaml"]

FROM gcr.io/distroless/static-debian12:latest-amd64 AS static

COPY --from=builder /app/crontab-go /crontab-go

ENV LOG_TIMESTAMP_FORMAT="2006-01-02T15:04:05.000Z"
ENV LOG_FORMAT="ansi"
ENV LOG_FILE=
ENV LOG_STDOUT=true
ENV LOG_LEVEL="debug"

ENV SHELL="static tag does not have a shell"
ENV SHELL_ARGS=""

ENV WEBSERVER_ADDRESS=
ENV WEBSERVER_PORT=
ENV WEBSERVER_USERNAME=
ENV WEBSERVER_PASSWORD=

WORKDIR /

ENTRYPOINT ["/crontab-go" ]
CMD ["-c","/config.yaml"]
