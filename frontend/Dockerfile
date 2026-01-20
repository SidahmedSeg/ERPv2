# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Accept build arguments for Next.js public env vars
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_BASE_DOMAIN
ARG NEXT_PUBLIC_GOOGLE_PLACES_API_KEY

# Set them as environment variables for the build
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_BASE_DOMAIN=$NEXT_PUBLIC_BASE_DOMAIN
ENV NEXT_PUBLIC_GOOGLE_PLACES_API_KEY=$NEXT_PUBLIC_GOOGLE_PLACES_API_KEY

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application (NEXT_PUBLIC_ vars will be baked in)
RUN npm run build

# Runtime stage
FROM node:20-alpine

WORKDIR /app

# Copy package files and install production dependencies only
COPY package*.json ./
RUN npm ci --production

# Copy built application from builder
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./

# Expose port
EXPOSE 3000

# Start the application
CMD ["npm", "start"]
