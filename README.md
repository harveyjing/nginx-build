# Nginx Binary Builder

This project contains a Dockerfile that builds a custom nginx binary from source code with HTTP/3 support.

## Features

- Builds nginx version 1.26.1 (configurable via ENV variable)
- Includes HTTP/3 (QUIC) support
- Uses BoringSSL (Google's SSL library with native QUIC support)
- Includes common modules like SSL, HTTP/2, and streaming
- Optimized for a minimal, production-ready binary

## Requirements

- Docker

## Usage

### Build the Docker image

```bash
docker build -t nginx-builder .
```

### Run the container to build nginx

```bash
docker run --name nginx-builder-container nginx-builder
```

### Extract the binary from the container

```bash
docker cp nginx-builder-container:/output/nginx ./nginx
```

### Verify the binary

```bash
chmod +x ./nginx
./nginx -v
```

## HTTP/3 Configuration

To use HTTP/3 in your nginx configuration, you need to:

1. Add the `http3` parameter to the listen directive:
   ```
   listen 443 ssl http3;
   ```

2. Enable QUIC and HTTP/3 transport for HTTPS:
   ```
   listen 443 quic reuseport;
   http3 on;
   ```

3. Add the Alt-Svc header to inform clients of HTTP/3 support:
   ```
   add_header Alt-Svc 'h3=":443"; ma=86400';
   ```

## Customization

You can modify the Dockerfile to:

1. Change the nginx version by updating the `NGINX_VERSION` environment variable
2. Add or remove compile options in the `./configure` step
3. Include additional dependencies or modules

## License

The nginx source code is subject to the [BSD-like license](https://nginx.org/LICENSE) used by the nginx project. 