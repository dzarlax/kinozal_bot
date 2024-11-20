# Stage 1: Use a lightweight image as base
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root

# Copy the pre-built Go binary from the build
COPY kinozalbot-linux .

RUN ls

# Ensure the binary has executable permissions
RUN chmod +x /root/kinozalbot-linux

# Command to run the binary when the container starts using the absolute path
CMD ["/root/kinozalbot-linux"]