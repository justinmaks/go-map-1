# Use official Go image
FROM golang:1.23-alpine

# Set environment variable (default is empty)
ENV IPINFO_TOKEN=""

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application code
COPY . .

# Build the application
RUN go build -o main .

# Expose port 8905
EXPOSE 8905

# Command to run the application
CMD ["./main"]
