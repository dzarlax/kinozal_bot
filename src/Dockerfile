# Stage 1: Use a lightweight image as base
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the pre-built Go binary from the build
COPY kinozalbot-linux .


# Command to run the binary when the container starts
CMD ["./kinozalbot-linux"]