# Nginx HTTP/3 Test Environment

This directory contains a test environment for the Nginx HTTP/3 build. It includes a Go service for testing various HTTP functionalities and Nginx as a reverse proxy.

## Project Structure

```
test/
├── README.md                 # This file
├── docker-compose.yml        # Docker compose configuration
├── hello-service/           # Go test service
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── nginx/                   # Nginx configuration
    ├── nginx.conf           # Main Nginx configuration
    └── conf.d/              # Modular configuration files
        ├── upstream.conf    # Upstream definitions
        ├── server.conf      # Server block configuration
        └── locations/       # Location-specific configurations
            └── api.conf     # API endpoints configuration
```

## Features

1. **Basic HTTP Endpoint**
   - Path: `/api/`
   - Returns: "Hello, World!"

2. **File Operations**
   - Path: `/api/file/{uuid}`
   - Supports both download (GET) and upload (POST)
   - UUID validation
   - Configurable file size for downloads
   - Streaming support for both operations

3. **JSON Endpoint**
   - Path: `/api/json`
   - Returns: JSON response

## Usage Examples

### File Download

```bash
# Download a 1MB random file (default size)
curl -O "http://localhost/api/file/123e4567-e89b-12d3-a456-426614174000"

# Download with custom size (5MB)
curl -O "http://localhost/api/file/123e4567-e89b-12d3-a456-426614174000?size=5242880"
```

### File Upload

```bash
# Create a test file
dd if=/dev/urandom of=test.bin bs=1M count=10

# Upload the file
curl -X POST -T test.bin http://localhost/api/file/123e4567-e89b-12d3-a456-426614174000
```

## Running the Environment

1. Start the environment:
   ```bash
   docker compose up --build
   ```

2. The following services will be available:
   - Nginx reverse proxy on ports 80 (HTTP) and 443 (HTTPS/HTTP3)
   - Go service internally on port 8080 (accessed through Nginx)

## Configuration

### Go Service
- Located in `hello-service/main.go`
- Handles file operations and basic HTTP endpoints
- Generates random data for downloads
- Validates UUIDs for file operations

### Nginx
- Uses the custom HTTP/3 build from the parent directory
- Configured to reverse proxy to the Go service
- Handles SSL/TLS termination
- Supports HTTP/3 protocol

## Testing Notes

1. File downloads generate random binary data
2. File uploads are currently processed but not stored (for testing purposes)
3. All endpoints are accessed through Nginx reverse proxy
4. SSL certificates are required for HTTPS/HTTP3 functionality

## Requirements

- Docker with Compose V2
- The parent directory's Nginx HTTP/3 build
- SSL certificates (for HTTPS/HTTP3) 