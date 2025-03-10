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
nginx-build/                    # Root project folder
├── Dockerfile                  # Main Nginx HTTP/3 build configuration
├── streaming/                  # File streaming application
│   ├── backend/                # Go backend service
│   │   ├── main.go            # Main application code
│   │   ├── go.mod             # Go module definition
│   │   ├── go.sum             # Go module checksums
│   │   └── data/              # Data directory for files
│   ├── frontend/              # Frontend static files
│   │   ├── index.html         # Main HTML file
│   │   ├── css/               # CSS styles
│   │   └── js/                # JavaScript code
│   ├── Dockerfile.backend     # Backend Docker configuration
│   ├── nginx.conf             # Default Nginx configuration
│   ├── nginx.conf.local       # Local development Nginx configuration
│   ├── nginx.conf.production  # Production Nginx configuration
│   ├── docker-compose.yml     # Base Docker Compose configuration
│   ├── docker-compose.local.yml # Local environment overrides
│   └── docker-compose.production.yml # Production environment overrides
└── certs/                      # SSL certificates for HTTPS/HTTP3
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
   mkdir -p streaming/data
   mkdir -p certs
   ```

3. Place your SSL certificates in the certs directory with the following structure (for production):
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
   cp -r /path/to/your/files streaming/data/
   ```

5. Build and start the containers using the appropriate environment:

   **For local development:**
   ```bash
   cd streaming
   docker-compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
   ```

   **For production deployment:**
   ```bash
   cd streaming
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

```
certbot certonly --standalone -d yourdomain.com
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
- `nginx.conf.production` for production environment

Then restart the appropriate environment:

```bash
# For local development
docker-compose -f docker-compose.yml -f docker-compose.local.yml restart nginx

# For production deployment
docker-compose -f docker-compose.yml -f docker-compose.production.yml restart nginx
```

## Scalability

For production deployment, consider:
- Using a managed container service (Kubernetes, ECS, etc.)
- Setting up a load balancer
- Implementing proper authentication
- Using object storage for files instead of local storage 