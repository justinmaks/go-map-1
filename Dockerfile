# Use official Go image
FROM golang:1.23-alpine

# Install GCC and SQLite development libraries
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Enable CGO for go-sqlite3
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Set environment variable for ipinfo token (default empty)
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
