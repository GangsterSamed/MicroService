FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /geo-service ./geo/cmd/

# Финальная стадия
FROM alpine:latest
WORKDIR /app
COPY --from=builder /geo-service /app/geo-service
COPY --from=builder /app/geo/docs /app/docs
HEALTHCHECK --interval=30s --timeout=3s \
  CMD grpc_health_probe -addr=:50051 || exit 1
CMD ["/app/geo-service"]