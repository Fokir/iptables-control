# Dockerfile for testing iptables functionality
FROM node:20-alpine

# Install iptables and required tools
RUN apk add --no-cache \
    iptables \
    iptables-legacy \
    bash \
    curl \
    jq

# Create iptables rules directory
RUN mkdir -p /etc/iptables

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy application source
COPY . .

# Build the application
RUN npm run build

# Create database directory
RUN mkdir -p /app/database

# Copy entrypoint script and fix line endings (Windows CRLF -> Unix LF)
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN sed -i 's/\r$//' /docker-entrypoint.sh && chmod +x /docker-entrypoint.sh

# Expose port
EXPOSE 3000

# Set environment variables
ENV BASIC_AUTH_USER=admin
ENV BASIC_AUTH_PASSWORD=admin

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["npm", "run", "start"]
