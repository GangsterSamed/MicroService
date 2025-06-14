FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./user/cmd/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /user-service /app/user-service
COPY --from=builder /app/user/migrations /app/migrations
EXPOSE 8080
CMD ["/app/user-service"]