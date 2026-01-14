#!/bin/bash

# ============================================================================
# Deployment Script for Bolsillo Claro Backend
# ============================================================================

set -e  # Exit on error

echo "🚀 Starting deployment..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo -e "${RED}❌ Error: .env.production not found${NC}"
    echo "Please create .env.production from .env.production.example"
    exit 1
fi

# Stop and remove old container if exists
echo -e "${YELLOW}🛑 Stopping old container...${NC}"
docker stop bolsillo-claro-backend 2>/dev/null || true
docker rm bolsillo-claro-backend 2>/dev/null || true

# Build new image
echo -e "${YELLOW}🔨 Building Docker image...${NC}"
docker build -t bolsillo-claro-backend:latest .

# Run new container
echo -e "${YELLOW}🚀 Starting new container...${NC}"
docker run -d \
  --name bolsillo-claro-backend \
  --restart unless-stopped \
  --add-host=host.docker.internal:host-gateway \
  --env-file .env.production \
  -p 8080:8080 \
  bolsillo-claro-backend:latest

# Wait for container to be healthy
echo -e "${YELLOW}⏳ Waiting for container to be healthy...${NC}"
sleep 5

# Check if container is running
if docker ps | grep -q bolsillo-claro-backend; then
    echo -e "${GREEN}✅ Deployment successful!${NC}"
    echo -e "${GREEN}Backend running on http://localhost:8080${NC}"
    echo ""
    echo "📊 Container logs:"
    docker logs bolsillo-claro-backend --tail 20
else
    echo -e "${RED}❌ Deployment failed!${NC}"
    echo "Container logs:"
    docker logs bolsillo-claro-backend
    exit 1
fi
