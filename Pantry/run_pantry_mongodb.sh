#!/bin/bash

# Build MongoDB Docker image
docker build -t pantry-mongodb .

# Run MongoDB container
docker run -d -p 27017:27017 --name pantry-mongodb-instance pantry-mongodb

# Get MongoDB container IP address
MONGODB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' pantry-mongodb-instance)

# Print MongoDB endpoint
echo "MongoDB endpoint: mongodb://${MONGODB_IP}:27017/"