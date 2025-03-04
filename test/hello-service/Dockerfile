FROM golang:1.21.6-alpine3.19 AS builder

WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/main .
COPY static/ /app/static/

# Create directory for file storage
RUN mkdir -p /data/files && \
    chmod 755 /data/files

EXPOSE 8080

CMD ["./main"] 