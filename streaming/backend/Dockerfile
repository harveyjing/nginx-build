# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy the source code
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    mkdir -p /app/data

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Create data directory for file storage
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"] 