# File Streaming Service with HTTP/3 Support

A web-based file management and streaming service with directory navigation, file downloads, and folder selection capabilities. This application leverages a custom-built Nginx with HTTP/3 support for optimal performance.

## Features

- Browse directories and files
- Select and download multiple files
- Download entire folders with their contents
- Responsive UI with file type icons
- Breadcrumb navigation
- HTTP/3 (QUIC) support for faster loading
- Secure HTTPS connections
- Environment-specific configurations (local/production)

## Architecture

- **Frontend**: Static HTML/CSS/JS served by custom Nginx with HTTP/3 support
- **Backend**: Go API for file operations
- **Docker**: Containerized deployment with multi-stage builds

## Directory Structure

```
nginx-build/                      # Root project folder
├── Dockerfile                    # Main Nginx HTTP/3 build configuration
├── setup-ssl.sh                 # SSL certificate generation script
├── streaming/                    # File streaming application
│   ├── backend/                  # Go backend service
│   │   ├── main.go              # Main application code
│   │   ├── go.mod              # Go module definition
│   │   ├── go.sum              # Go module checksums
│   │   └── data/               # Data directory for files
│   ├── frontend/                # Frontend static files
│   │   ├── index.html          # Main HTML file
│   │   ├── css/                # CSS styles
│   │   └── js/                 # JavaScript code
│   ├── Dockerfile.backend       # Backend Docker configuration
│   ├── nginx.conf              # Default Nginx configuration
│   ├── nginx.conf.local        # Local development Nginx configuration
│   ├── nginx.conf.production   # Production Nginx configuration template
│   ├── default.conf            # Default server configuration
│   ├── default.conf.local      # Local server configuration
│   ├── default.conf.production # Production server configuration
│   ├── default.conf.production.template # Template for production server config
│   ├── docker-compose.yml      # Base Docker Compose configuration
│   ├── docker-compose.local.yml # Local environment overrides
│   └── docker-compose.production.yml # Production environment overrides
└── certs/                       # SSL certificates for HTTPS/HTTP3
    └── live/                    # Let's Encrypt live certificates
        └── example.com/         # Domain-specific certificates
            ├── fullchain.pem    # Full certificate chain
            ├── privkey.pem     # Private key
            └── chain.pem       # Certificate chain
```

## Environment Configuration

The service supports two environments:

### Local Development Environment
- Optimized for testing on localhost
- Simplified HTTP configuration without SSL requirement
- Configured via docker-compose.local.yml

### Production Environment
- Optimized for deployment on example.com
- Full HTTPS with HTTP/3 support
- Enhanced security headers
- Caching for static assets
- Configured via docker-compose.production.yml

## Deployment with Docker

### Prerequisites

- Docker
- Docker Compose
- SSL certificates for HTTPS/HTTP3 (required for production, optional for local)

### Building and Running

1. Clone the repository:
   ```
   git clone <repository-url>
   cd nginx-build
   ```

2. Create data directory if it doesn't exist:
   ```
   mkdir -p streaming/backend/data
   mkdir -p certs
   ```

3. Generate certificates by `setup-ssl.sh` script at root folder. Those files will be mounted inside docker container like blow (for production):
   ```
   certs/
   └── live/
       └── example.com/  # Replace with your domain
           ├── fullchain.pem
           ├── privkey.pem
           └── chain.pem
   ```

4. Add some test files to the data directory (optional):
   ```
   cp -r /path/to/your/files streaming/backend/data/
   ```

5. Build and start the containers using the appropriate environment:

   **For local development:**
   ```bash
   cd streaming
   docker-compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
   ```

   **For production deployment:**
   Before deploying to production, you need to configure your domain:

   1. Replace the PRIMARY_DOMAIN placeholder in the template:
      ```bash
      # Replace example.com with your actual domain
      sed "s/PRIMARY_DOMAIN/example.com/g" default.conf.production.template > default.conf.production
      ```
   2. Verify the configuration:
      ```bash
      grep "server_name" default.conf.production
      # Should output: server_name example.com www.example.com;
      ```
   3. Ensure your SSL certificates are in the root `cert` folder:
   4. Then start your server by:
      ```bash
      docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d --build
      ```
6. Access the application:
   - Local environment: http://localhost
   - Production environment: https://example.com (or http://example.com which redirects)
   - Direct API Access: http://localhost:8080/api

### Stopping the Service

```bash
docker-compose down
```

## SSL Certificates

For HTTPS and HTTP/3 to work in production, you need valid SSL certificates. You can use Let's Encrypt to generate free SSL certificates with the following command:

```sh
$PROJECT_ROOT/setup-ssl.sh
```

## HTTP/3 Benefits

HTTP/3 offers several advantages over HTTP/2:

- **Faster connection establishment** - Reduces latency with 0-RTT handshakes
- **Improved performance on unreliable networks** - Better handling of packet loss
- **Elimination of head-of-line blocking** - Multiple streams operate independently
- **Connection migration** - Maintains connections when changing networks

## API Endpoints

- `GET /api/files?path=<directory>` - List files in directory
- `GET /api/download?files=<file1>&files=<file2>` - Download files
- `GET /api/health` - Health check

## Configuration

### Modifying Backend

Edit `backend/main.go` to add features, then rebuild:
```
docker-compose up -d --build backend
```

### Modifying Frontend

Edit files in the `frontend` directory. Changes will be reflected immediately since they're mounted into the container.

### Changing Nginx Configuration

Edit the appropriate configuration file:
- `nginx.conf.local` for local environment
- `nginx.conf.production` for production environment (Generated by the template)

Then restart the appropriate environment:

```bash
# For local development
docker-compose -f docker-compose.yml -f docker-compose.local.yml restart nginx

# For production deployment
docker-compose -f docker-compose.yml -f docker-compose.production.yml restart nginx
```
