FROM golang:1.9.2-alpine3.6 AS builder
RUN apk add --no-cache -q gcc musl-dev
WORKDIR /build
COPY ts3_discord_bot.go .
RUN CC=$(which musl-gcc) CGO_ENABLED=0 GOOS=linux go build -a\
 -ldflags '-w -linkmode external -extldflags "-static"'\
 -o ts3_discord_bot

FROM alpine:3.6
RUN apk add --no-cache -q ca-certificates tzdata
WORKDIR /app
COPY wait-for .
COPY --from=builder /build/ts3_discord_bot .
CMD ["./ts3_discord_bot"]
