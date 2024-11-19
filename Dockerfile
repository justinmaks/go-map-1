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

# Set build argument for port
ARG GO_MAP_PORT=8905
ENV GO_MAP_PORT=${GO_MAP_PORT}

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

# Expose port
EXPOSE ${GO_MAP_PORT}

# Command to run the application
CMD ["./main"]