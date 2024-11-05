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

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create and set permissions for torrent_files directory
RUN mkdir torrent_files && chown appuser:appgroup torrent_files

# Copy built application from builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/node.js ./node.js
COPY --from=builder /app/package*.json ./

# Set environment variables
ENV NODE_ENV=production \
    TZ=UTC

# Switch to non-root user
USER appuser

# Set volume for torrent files
VOLUME ["/app/torrent_files"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD node -e "const http=require('http');const options={timeout:2000};const req=http.request('http://localhost:8080/health',options,(res)=>{if(res.statusCode==200){process.exit(0)}process.exit(1)});req.on('error',()=>process.exit(1));req.end()"

# Start the bot
CMD ["node", "node.js"]