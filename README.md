# Nginx HTTP/3 Docker Image Builder

This project contains a Dockerfile that builds a custom nginx Docker image with HTTP/3 support, along with a comprehensive test environment.

## Features

- Builds nginx version 1.26.1 (configurable via ENV variable)
- Includes HTTP/3 (QUIC) support
- Uses BoringSSL (Google's SSL library with native QUIC support)
- Includes common modules like SSL, HTTP/2, and streaming
- Optimized for a minimal, production-ready image
- Includes a test environment with:
  - Go service for testing HTTP functionalities
  - WebSocket support
  - File upload/download capabilities
  - Traffic analysis tools

## Project Structure

```
.
├── Dockerfile                    # Main nginx image build configuration
├── README.md                     # This file
├── setup-ssl.sh                  # SSL certificate setup script
├── default.conf                  # Default nginx configuration
├── streaming/                    # File streaming application
│   ├── backend/                 # Go backend service
│   │   ├── main.go             # Main application code
│   │   ├── go.mod              # Go module definition
│   │   ├── go.sum              # Go module checksums
│   │   ├── Dockerfile          # Backend Docker configuration
│   │   └── data/               # Data directory for files
│   ├── frontend/               # Frontend static files
│   │   ├── index.html         # Main HTML file
│   │   ├── css/               # CSS styles
│   │   └── js/                # JavaScript code
│   ├── nginx.conf             # Default Nginx configuration
│   ├── nginx.conf.local       # Local development Nginx config
│   ├── nginx.conf.production  # Production Nginx config
│   ├── default.conf           # Default server configuration
│   ├── default.conf.local     # Local server configuration
│   ├── default.conf.production # Production server config
│   ├── default.conf.production.template # Template for production
│   ├── docker-compose.yml     # Base Docker Compose config
│   ├── docker-compose.local.yml # Local environment overrides
│   └── docker-compose.production.yml # Production overrides
├── test/                        # Test environment
│   ├── README.md               # Test environment documentation
│   ├── docker-compose.yml      # Test services configuration
│   ├── docker-compose.local.yml # Local test overrides
│   ├── server.conf.local      # Local test server config
│   ├── hello-service/         # Go test service
│   │   ├── main.go           # Test service code
│   │   └── Dockerfile        # Test service build
│   └── nginx/                # Nginx test configuration
│       ├── nginx.conf        # Test nginx configuration
│       └── conf.d/           # Test server configurations
└── certs/                      # SSL certificates directory
    └── live/                   # Let's Encrypt certificates
        └── example.com/        # Domain-specific certs
            ├── fullchain.pem   # Full certificate chain
            ├── privkey.pem     # Private key
            └── chain.pem       # Certificate chain
```

## Requirements

- Docker
- Docker Compose V2 (for test environment)
- SSL certificates (for HTTPS/HTTP3 testing)

## SSL Certificate Setup

The project includes a `setup-ssl.sh` script that automates SSL certificate generation using Let's Encrypt. The script handles:
- Certificate generation for both domain.com and www.domain.com
- Proper certificate placement for nginx
- Automatic renewal setup
- Container restart with new certificates

### Using setup-ssl.sh

```bash
# Basic usage with just domain
./setup-ssl.sh -d example.com

# With custom email
./setup-ssl.sh -d example.com -e admin@example.com

# With custom container and image names
./setup-ssl.sh -d example.com -c my-nginx -i my-http3-image
```

Available options:
- `-d domain`: Domain name (required, e.g., example.com)
- `-e email`: Email for Let's Encrypt notifications (default: admin@domain)
- `-c name`: Container name (default: nginx-https-server)
- `-i name`: Image name (default: nginx-http3)
- `-h`: Show help message

The script will:
1. Stop any existing nginx container
2. Generate certificates using Let's Encrypt
3. Create proper symbolic links
4. Start nginx with the new certificates
5. Set up monthly automatic renewal

### Certificate Location

After running the script, certificates will be available at:
```
certs/live/your-domain.com/
├── fullchain.pem  # Full certificate chain
├── privkey.pem    # Private key
└── chain.pem      # Certificate chain
```

## Usage

### Build the Docker image

```bash
docker build -t nginx-http3 .
```

### Run the container

```bash
docker run -d \
  --name nginx-http3-server \
  -p 80:80 \
  -p 443:443 \
  -p 443:443/udp \
  -v /path/to/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v /path/to/conf.d:/etc/nginx/conf.d:ro \
  -v /path/to/certs:/etc/nginx/certs:ro \
  nginx-http3
```

### Verify the container

```bash
# Check container status
docker ps

# Check nginx version and configuration
docker exec nginx-http3-server nginx -v
docker exec nginx-http3-server nginx -t
```

## Test Environment

The project includes a comprehensive test environment located in the `test/` directory. To use it:

1. Navigate to the test directory:
   ```bash
   cd test
   ```

2. Start the test environment:
   ```bash
   docker compose up --build
   ```

3. Access the test services:
   - HTTP: http://localhost
   - HTTPS/HTTP3: https://localhost
   - WebSocket test page: http://localhost/ws-test

For detailed information about the test environment, including available endpoints and traffic analysis, see [test/README.md](test/README.md).

## HTTP/3 Configuration

To enable HTTP/3 support in nginx, you need to properly configure your server blocks. HTTP/3 requires HTTPS and QUIC protocol support. Here's the correct configuration:

```nginx
server {
    # Standard HTTP port - redirect to HTTPS
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}

server {
    # Configure HTTP/3
    listen 443 ssl;                # Standard HTTPS
    http2 on;                      # Enable http2
    listen 443 quic reuseport;     # UDP port for QUIC+HTTP/3
    server_name example.com;
    
    # SSL certificates
    ssl_certificate     /etc/nginx/certs/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/live/example.com/privkey.pem;
    
    # Add Alt-Svc header to inform clients about HTTP/3 support
    add_header Alt-Svc 'h3=":443"; ma=86400';
    
    # SSL settings
    ssl_protocols TLSv1.3;         # TLSv1.3 preferred for HTTP/3
    
    # Your server configuration
    location / {
        root /usr/share/nginx/html;
        index index.html;
    }
}
```

Key points to understand:

1. **Dual Listening**:
   - Standard TLS over TCP: `listen 443 ssl;`
   - Enable HTTP/2: `http2 on;`

2. **HTTP/3 Enabling**:
   - QUIC+HTTP/3 over UDP: `listen 443 quic reuseport;`

3. **Alt-Svc Header**:
   - The `Alt-Svc` header tells clients the server supports HTTP/3
   - The value `h3=":443"; ma=86400` means HTTP/3 is available on port 443 with a max age of 1 day

4. **TLSv1.3 Recommendation**:
   - HTTP/3 works best with TLSv1.3 for optimal performance

## Customization

You can modify the Dockerfile to:

1. Change the nginx version by updating the `NGINX_VERSION` environment variable
2. Add or remove compile options in the `./configure` step
3. Include additional dependencies or modules
4. Modify build optimization flags

## Testing HTTP/3 Support

To verify HTTP/3 support:

1. Use curl with HTTP/3 support:
   ```bash
   curl --http3 https://localhost/api/
   ```

2. Check the Alt-Svc header in responses:
   ```bash
   curl -I https://localhost/api/
   ```

3. Use Chrome or Edge browser to test HTTP/3 connections (they support HTTP/3 by default)

## License

The nginx source code is subject to the [BSD-like license](https://nginx.org/LICENSE) used by the nginx project.