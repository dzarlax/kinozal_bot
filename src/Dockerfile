# Use a base image for building (Alpine is a lightweight base image)
FROM golang:1.23-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download all the dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go app
RUN GOOS=linux GOARCH=amd64 go build -o kinozal-bot-linux main.go

# Start a new stage from scratch (smaller image)
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary from the previous stage
COPY --from=builder /app/kinozal-bot-linux .

# Expose the port (optional)
EXPOSE 8080

# Run the binary
CMD ["./kinozal-bot-linux"]