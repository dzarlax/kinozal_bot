# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Production stage
FROM node:20-alpine


# Set working directory
WORKDIR /app

# Create folder for torrents
RUN mkdir torrents

# Copy application files from builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/node.js ./
COPY --from=builder /app/config.js ./
COPY --from=builder /app/errors.js ./
COPY --from=builder /app/errorHandler.js ./
COPY --from=builder /app/fileutils.js ./
COPY --from=builder /app/logger.js ./
COPY --from=builder /app/middleware.js ./
COPY --from=builder /app/userManagement.js ./
COPY --from=builder /app/menu.js ./
COPY --from=builder /app/package*.json ./



# Set environment variables
ENV NODE_ENV=production \
    TZ=UTC


# Health check using transmission RPC
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:${TRANSMISSION_PORT:-9091}/transmission/rpc || exit 1

# Start the bot
CMD ["node", "node.js"]