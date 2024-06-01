FROM golang:latest as BUILDER
COPY ./ /app
WORKDIR /app
RUN go build -o crontab-go

FROM gcr.io/distroless/static-debian12:noneroot
COPY BUILDER:/app/crontab-go /bin/crontab-go
ENTRYPOINT ["/bin/crontab-go"]
