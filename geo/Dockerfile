FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /geo-service ./geo/cmd/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /geo-service /app/geo-service
EXPOSE 8080
CMD ["/app/geo-service"]