# Use an official Golang image as the base
FROM golang:1.23-alpine

# Install required build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Enable CGO and build the application
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -o main .

# Expose the port the app listens on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
