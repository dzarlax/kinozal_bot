# Stage 1: Use a lightweight image as base
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the pre-built Go binary from the build
COPY kinozal-bot-linux .

# Optionally expose port (if required)
EXPOSE 8080

# Command to run the binary when the container starts
CMD ["./kinozal-bot-linux"]