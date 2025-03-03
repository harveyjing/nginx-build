# Nginx Binary Builder

This project contains a Dockerfile that builds a custom Nginx image from source code with HTTP/3 support.

## Features

- Builds nginx version 1.26.3 (configurable via ENV variable)
- Includes HTTP/3 (QUIC) support
- Uses BoringSSL (Google's SSL library with native QUIC support)
- Includes common modules like SSL, HTTP/2, and streaming
- Optimized for a minimal, production-ready image

## Requirements

- Docker

## Usage

### Build the Docker image

```bash
docker build -t nginx-http3 .
```

### Run the container to build nginx

```bash
docker run --rm -d --name nginx-http3-server -p 80:80 -p 443:443 -p 443:443/udp nginx-http3
```
