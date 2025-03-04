FROM golang:1.21.6-alpine3.19 AS builder

WORKDIR /app
COPY . .
RUN go build -o hello-service

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/hello-service .

EXPOSE 8080

CMD ["./hello-service"] 