FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /proxy-service ./proxy/cmd/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /proxy-service /app/proxy-service
EXPOSE 8080 
CMD ["/app/proxy-service"]