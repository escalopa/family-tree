# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build application
RUN npm run build

# Development stage - for hot reload
FROM node:20-alpine AS development

WORKDIR /app

# Install dependencies
RUN apk add --no-cache curl

# Copy package files
COPY package.json package-lock.json ./
RUN npm ci

# Start dev server
CMD ["npm", "run", "dev"]

# Production stage - serve with nginx
FROM nginx:alpine AS production

# Copy built files from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
