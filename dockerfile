# Base image
FROM node:20

# Create and set the working directory
WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy the source code
COPY . .

# Set environment variables file path
ENV NODE_ENV=production

# Start the bot
CMD ["node", "node.js"]