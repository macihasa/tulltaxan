# Use a specific Go version
FROM golang:1.23.4

# Set working directory in the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . ./

# Expose the application port
EXPOSE 8080

# Set the default command to run
CMD ["make", "dev"]
