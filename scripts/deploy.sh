#!/bin/bash

# Root Server Deployment Script

set -e

echo "Deploying Root Server..."

# Build the binary
echo "Building..."
make build

# Run tests
echo "Running tests..."
make test

# Build Docker image
echo "Building Docker image..."
docker build -t root-server:$(git rev-parse --short HEAD) .
docker tag root-server:$(git rev-parse --short HEAD) root-server:latest

# Push to registry (customize for your registry)
# docker push root-server:latest

echo "Deployment complete!"
