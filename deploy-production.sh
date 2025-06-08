#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables
if [ -f .env ]; then
    print_status "Loading environment variables from .env file..."
    set -a
    source .env
    set +a
else
    print_error ".env file not found! Please create it with required variables."
    exit 1
fi

# Check required environment variables
if [ -z "$DOMAIN_NAME" ]; then
    print_error "DOMAIN_NAME is not set in .env file"
    exit 1
fi

print_status "Starting production deployment for domain: $DOMAIN_NAME"

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p ./certbot/conf
mkdir -p ./certbot/www
mkdir -p ./nginx

# Check if SSL certificates already exist
if [ -f "./certbot/conf/live/$DOMAIN_NAME/fullchain.pem" ]; then
    print_success "SSL certificates already exist for $DOMAIN_NAME"
    USE_SSL=true
else
    print_warning "No SSL certificates found. Will obtain them after nginx starts."
    USE_SSL=false
fi

# Create nginx configuration files
print_status "Creating nginx configuration templates..."

# Copy the templates to nginx directory (these should be created as separate files)
if [ ! -f "./nginx/http-template" ] || [ ! -f "./nginx/https-template" ]; then
    print_error "Nginx templates not found! Please ensure http-template and https-template exist in ./nginx/ directory"
    exit 1
fi

# Step 1: Start the application without SSL
print_status "Step 1: Starting application in HTTP-only mode..."
export SSL_MODE=development
docker compose up -d ragbot db nginx

# Wait for services to be healthy
print_status "Waiting for services to become healthy..."
sleep 10

# Check if nginx is responding
for i in {1..30}; do
    if curl -f -s "http://localhost/health" > /dev/null 2>&1; then
        print_success "Nginx is responding on HTTP"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "Nginx failed to start properly"
        docker compose logs nginx
        exit 1
    fi
    sleep 2
done

# Step 2: Obtain SSL certificates if they don't exist
if [ "$USE_SSL" = false ]; then
    print_status "Step 2: Obtaining SSL certificates..."

    # Set certbot environment variables
    export SSL_EMAIL="${SSL_EMAIL:-admin@$DOMAIN_NAME}"
    export CERTBOT_STAGING="${CERTBOT_STAGING:-}"  # Remove --staging for production

    # Run certbot to obtain certificates using a one-time container
    docker compose run --rm certbot certbot certonly --webroot -w /var/www/certbot \
        --email "$SSL_EMAIL" \
        -d "$DOMAIN_NAME" \
        --rsa-key-size 4096 \
        --agree-tos \
        --non-interactive \
        $CERTBOT_STAGING

    if [ $? -eq 0 ]; then
        print_success "SSL certificates obtained successfully!"
    else
        print_error "Failed to obtain SSL certificates"
        print_status "Checking certbot logs..."
        docker compose logs certbot
        exit 1
    fi
fi

# Step 3: Switch to HTTPS configuration
print_status "Step 3: Switching to HTTPS configuration..."
export SSL_MODE=production

# Restart nginx with SSL configuration
docker compose up -d --force-recreate nginx

# Wait for nginx to restart with SSL
sleep 5

# Verify HTTPS is working
for i in {1..30}; do
    if curl -f -s -k "https://localhost/health" > /dev/null 2>&1; then
        print_success "Nginx is responding on HTTPS"
        break
    fi
    if [ $i -eq 30 ]; then
        print_warning "HTTPS verification failed, but this might be normal with self-signed certificates"
        break
    fi
    sleep 2
done

# Step 4: Start certbot renewal service
print_status "Step 4: Starting SSL certificate renewal service..."
docker compose --profile ssl up -d certbot

print_success "Production deployment completed!"

print_status "Service status:"
docker compose ps

print_status "To verify your deployment:"
echo "  - HTTP:  http://$DOMAIN_NAME"
echo "  - HTTPS: https://$DOMAIN_NAME"

print_status "To view logs:"
echo "  docker compose logs -f [service_name]"

print_status "To stop all services:"
echo "  docker compose --profile ssl down"

print_warning "Important: Make sure your domain DNS A record points to this server's IP address!"

# Optional: Test the domain resolution
if command -v dig &> /dev/null; then
    print_status "Testing DNS resolution for $DOMAIN_NAME..."
    RESOLVED_IP=$(dig +short $DOMAIN_NAME | tail -n1)
    if [ -n "$RESOLVED_IP" ]; then
        print_success "Domain resolves to: $RESOLVED_IP"
    else
        print_warning "Could not resolve domain. Please check your DNS settings."
    fi
fi