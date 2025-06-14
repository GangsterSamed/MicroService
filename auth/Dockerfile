FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./auth/cmd/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /auth-service /app/auth-service
EXPOSE 8080
CMD ["/app/auth-service"]