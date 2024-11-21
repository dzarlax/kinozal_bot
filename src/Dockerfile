FROM alpine:latest as alpine
RUN apk add --no-cache libc6-compat

# Stage 1: Use a lightweight image as base
FROM scratch

# Install necessary runtime dependencies


# Set the working directory inside the container
WORKDIR /root

# Copy the pre-built Go binary from the build
COPY kinozalbot-linux .

# Ensure the binary has executable permissions
RUN chmod +x /root/kinozalbot-linux && \
    # Print file type information for debugging
    file /root/kinozalbot-linux && \
    # Print dynamic library dependencies
    ldd /root/kinozalbot-linux || true && \
    # Add the current directory to PATH
    export PATH="/root:$PATH"

# Command to run the binary when the container starts using the absolute path
CMD ["/root/kinozalbot-linux"]