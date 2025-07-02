# monitoring-ui/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod* ./
RUN go get k8s.io/client-go@v0.28.3 k8s.io/api@v0.28.3 k8s.io/apimachinery@v0.28.3
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/monitoring-ui .

FROM scratch
COPY --from=builder /app/monitoring-ui /monitoring-ui
COPY index.html /index.html
EXPOSE 8080
ENTRYPOINT ["/monitoring-ui"]