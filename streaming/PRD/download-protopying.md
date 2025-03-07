# Download Propotyping

The Architecture is the serevr which is golang is behind the Nginx. And it's a web service has a web UI like a list display all the files which can be download in a folder. 

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│              │     │              │     │   Golang     │
│    Client    ├────►│    Nginx     ├────►│   Server     │
│   Browser    │     │  (Reverse    │     │  (Backend)   │
│              │     │   Proxy)     │     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                 │
                                                 │
                                          ┌──────▼───────┐
                                          │   File       │
                                          │   Storage    │
                                          │              │
                                          └──────────────┘

Flow:
1. Client accesses web UI through browser
2. Nginx serves static content and proxies API requests
3. Golang server handles file operations and streaming
4. Files are served from storage through optimized streaming
```

## Features

1. Users can select multi files by togle the checkbox in front of the list items;
2. Users can download the multi files in a single request;
3. The performance is critical because users could download huge files;

## Techonical Details

## Implementing

### 1. Project Setup
1. Create project structure:
   ```
   .
   ├── docker-compose.yml
   ├── nginx
   │   ├── Dockerfile
   │   └── conf
   │       └── nginx.conf
   ├── backend
   │   ├── Dockerfile
   │   ├── main.go
   │   ├── go.mod
   │   └── go.sum
   └── frontend
       ├── index.html
       ├── css
       └── js
   ```

2. Initialize Go module and install dependencies:
   ```bash
   cd backend
   go mod init download-service
   go get -u github.com/gin-gonic/gin
   ```

### 2. Backend Implementation (Golang)
1. Create REST API endpoints:
   - `GET /api/files` - List all available files
   - `GET /api/download` - Handle single/multiple file downloads
   - `GET /api/health` - Health check endpoint

2. Implement file operations:
   - Directory scanning
   - File streaming with proper chunking
   - Multi-file zip streaming
   - Progress tracking

3. Add error handling and logging:
   - Request validation
   - Error responses
   - Access logging
   - Performance metrics

### 3. Frontend Implementation
1. Create responsive UI components:
   - File list view with checkboxes
   - Download progress indicator
   - File size and type information
   - Error message display

2. Implement JavaScript functionality:
   - File selection handling
   - Download progress tracking
   - API integration
   - Error handling

### 4. Nginx Configuration
1. Set up reverse proxy:
   ```nginx
   location /api/ {
       proxy_pass http://backend:8080/;
       proxy_buffering off;
       proxy_request_buffering off;
   }
   ```

2. Configure for large file handling:
   - Enable streaming
   - Set appropriate buffer sizes
   - Configure timeouts

### 5. Docker Setup
1. Create Dockerfiles for each service
2. Configure docker-compose.yml:
   - Nginx service
   - Backend service
   - Volume mappings
   - Network settings

### 6. Testing
1. Unit tests for backend:
   - File operations
   - API endpoints
   - Error handling

2. Integration tests:
   - End-to-end download flow
   - Multi-file download
   - Large file handling

3. Performance testing:
   - Concurrent downloads
   - Large file streaming
   - Memory usage monitoring

### 7. Monitoring and Optimization
1. Add monitoring:
   - Download speeds
   - Server resource usage
   - Error rates

2. Implement optimizations:
   - File caching
   - Stream compression
   - Connection pooling

### 8. Security Measures
1. Implement basic security:
   - Rate limiting
   - File type validation
   - Maximum file size limits
   - Access controls

## Requirements

1. Use Docker Compose V2;
2. Nginx is build from the project root folder;
3. Struct the project under the `./streaming` folder;

## Potential Issues

1. The server have used streaming api. But it seems the compression speed is not quite fast. Now it's about 40MB/s. But if I download a single without zip, the speed can be as fast as 400Mb/s;
   1. Which is solved by using `Store` compress option;
2. Seems that the final zip file's size can't be pre-calculated because some dynamic metadata. Need more research;