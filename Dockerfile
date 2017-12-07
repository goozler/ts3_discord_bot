FROM golang:1.9.2-alpine3.6 AS builder
RUN apk add --no-cache -q gcc musl-dev
WORKDIR /build
COPY ts3_discord_bot.go .
RUN CC=$(which musl-gcc) CGO_ENABLED=0 GOOS=linux go build -a\
 -ldflags '-w -linkmode external -extldflags "-static"'\
 -o ts3_discord_bot

FROM alpine:3.7
ARG BUILD_DATE
ARG VCS_REF
LABEL maintainer="Krutov Alexander <goozler@mail.ru>" \
      org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-url="https://github.com/goozler/ts3_discord_bot.git" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.schema-version="1.0.0-rc.1" \
      org.label-schema.description="Send a notification to Discord when someone \
has connected or disconnected to a TeamSpeak3 server"

RUN apk add --no-cache -q ca-certificates tzdata
WORKDIR /app
COPY wait-for .
COPY --from=builder /build/ts3_discord_bot .
CMD ["./ts3_discord_bot"]
