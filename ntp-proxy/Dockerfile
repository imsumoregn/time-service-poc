# ntp-proxy/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/ntp-proxy .

FROM alpine:3.18
RUN apk update && apk add --no-cache chrony
COPY --from=builder /app/ntp-proxy /usr/local/bin/ntp-proxy
COPY start.sh /start.sh
RUN chmod +x /start.sh
EXPOSE 8080
CMD ["/start.sh"]