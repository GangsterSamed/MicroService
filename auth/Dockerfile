FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./auth/cmd/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /auth-service /app/auth-service
COPY --from=builder /app/auth/docs /app/docs
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -q -O /dev/null http://localhost:8081/health || exit 1
CMD ["/app/auth-service"]