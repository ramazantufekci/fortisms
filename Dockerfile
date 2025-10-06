FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN go mod init helper
RUN CGO_ENABLED=0 go build -o /fortisms

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /fortisms /app/fortisms
COPY config.json /app/config.json
ENTRYPOINT ["/app/fortisms"]
