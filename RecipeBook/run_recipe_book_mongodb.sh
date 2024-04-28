#!/bin/bash

# Build MongoDB Docker image
docker build -t recipe-book-mongodb .

# Run MongoDB container
docker run -d -p 27017:27017 --name recipe-book-mongodb-instance recipe-book-mongodb

# Get MongoDB container IP address
MONGODB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' recipe-book-mongodb-instance)

# Print MongoDB endpoint
echo "MongoDB endpoint: mongodb://${MONGODB_IP}:27017/"