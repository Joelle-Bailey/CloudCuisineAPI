# Use the official Go image as a base
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go binary inside the container
RUN go build -o usermanagement .

# Expose the port the application runs on
EXPOSE 8080

# Command to run the executable
CMD ["./usermanagement"]
