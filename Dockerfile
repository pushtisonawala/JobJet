# Use the official Go image (lightweight Alpine Linux)
FROM golang:1.25-alpine

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the project files
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8000 to the host
EXPOSE 8000

# Run the compiled binary when container starts
CMD ["./main"]
