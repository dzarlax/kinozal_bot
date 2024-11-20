# Stage 1: Build the Go binary
FROM golang:1.23-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY ./src/go.mod ./src/go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the application source code into the container
COPY ./src/ .

# Build the Go app for Linux
RUN GOOS=linux GOARCH=amd64 go build -o kinozal-bot-linux main.go

# Stage 2: Create a smaller image for runtime
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/kinozal-bot-linux .

# Optionally expose port (e.g., for HTTP server)
EXPOSE 8080

# Command to run the binary when the container starts
CMD ["./kinozal-bot-linux"]