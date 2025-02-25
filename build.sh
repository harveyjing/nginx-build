#!/bin/bash
set -e

echo "==== Building nginx binary ===="

# Build the Docker image
echo "Building Docker image..."
docker build -t nginx-builder .

# Run the container to build nginx
echo "Running container to build nginx..."
docker run --name nginx-builder-container nginx-builder

# Extract the binary from the container
echo "Extracting nginx binary..."
docker cp nginx-builder-container:/output/nginx ./nginx

# Make the binary executable
chmod +x ./nginx

# Clean up the container
echo "Cleaning up container..."
docker rm nginx-builder-container

echo "==== Build completed successfully ===="
echo "The nginx binary is available at: $(pwd)/nginx"
echo "Nginx version info:"
./nginx -v 