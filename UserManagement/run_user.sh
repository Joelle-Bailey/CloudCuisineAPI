#!/bin/bash

# Define Docker image and container names
DOCKER_IMAGE_NAME="user-management"
DOCKER_CONTAINER_NAME="user-management-instance"

# Remove existing Docker container if it exists
docker rm -f "$DOCKER_CONTAINER_NAME"

# Build Go Docker image
docker build -t "$DOCKER_IMAGE_NAME" .

# Run Go Docker container
docker run -d -e SPOONACULAR_API_KEY=2cb224078c8346dc98e48581d25d0788 -p 8080:8080 --name "$DOCKER_CONTAINER_NAME" "$DOCKER_IMAGE_NAME"

# Print endpoint
echo "Go application endpoint: http://localhost:8080/"
