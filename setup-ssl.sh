#!/bin/bash
set -e

# Default values
CONTAINER_NAME="nginx-https-server"
IMAGE_NAME="nginx-http3"

# Help function
usage() {
    echo "Usage: $0 -d domain [-e email] [-c container_name] [-i image_name]"
    echo
    echo "Options:"
    echo "  -d domain        Domain name (required, e.g., example.com)"
    echo "  -e email        Email address for Let's Encrypt notifications (default: admin@domain)"
    echo "  -c name         Container name (default: nginx-https-server)"
    echo "  -i name         Image name (default: nginx-http3)"
    echo "  -h             Show this help message"
    exit 1
}

# Parse command line arguments
while getopts "d:e:c:i:h" opt; do
    case $opt in
        d)
            DOMAIN="$OPTARG"
            ;;
        e)
            EMAIL="$OPTARG"
            ;;
        c)
            CONTAINER_NAME="$OPTARG"
            ;;
        i)
            IMAGE_NAME="$OPTARG"
            ;;
        h)
            usage
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            usage
            ;;
    esac
done

# Check if domain is provided
if [ -z "$DOMAIN" ]; then
    echo "Error: Domain name is required"
    usage
fi

# Set default email if not provided
if [ -z "$EMAIL" ]; then
    EMAIL="admin@$DOMAIN"
fi

# Set www subdomain
PRIMARY_DOMAIN="www.$DOMAIN"

echo "Generating SSL certificates for:"
echo "Domain: $DOMAIN"
echo "WWW Domain: $PRIMARY_DOMAIN"
echo "Email: $EMAIL"
echo

# Create directory for certbot
mkdir -p ./certs/certbot

# Stop any running nginx container
docker stop $CONTAINER_NAME || true
docker rm $CONTAINER_NAME || true

# Get SSL certificate using certbot
docker run -it --rm \
  -v "$(pwd)/certs:/etc/letsencrypt" \
  -v "$(pwd)/certs/certbot:/var/lib/letsencrypt" \
  -p 80:80 \
  certbot/certbot certonly \
  --standalone \
  --preferred-challenges http \
  --email $EMAIL \
  --agree-tos \
  --no-eff-email \
  -d $DOMAIN \
  -d $PRIMARY_DOMAIN

# Create symbolic links for nginx
mkdir -p ./certs/live/$DOMAIN
ln -sf ../../archive/$DOMAIN/fullchain1.pem ./certs/live/$DOMAIN/fullchain.pem
ln -sf ../../archive/$DOMAIN/privkey1.pem ./certs/live/$DOMAIN/privkey.pem
ln -sf ../../archive/$DOMAIN/chain1.pem ./certs/live/$DOMAIN/chain.pem

# Start nginx with the new certificates
docker run -d --name $CONTAINER_NAME \
  -p 80:80 \
  -p 443:443 \
  -p 443:443/udp \
  -v $(pwd)/certs:/etc/nginx/certs:ro \
  -v $(pwd)/default.conf:/etc/nginx/conf.d/default.conf:ro \
  $IMAGE_NAME

# Set up automatic renewal
echo "0 0 1 * * docker run --rm -v $(pwd)/certs:/etc/letsencrypt -v $(pwd)/certs/certbot:/var/lib/letsencrypt certbot/certbot renew --quiet && docker restart $CONTAINER_NAME" | sudo tee -a /var/spool/cron/crontabs/root

echo "SSL certificates have been obtained and nginx has been started!"
echo "Automatic renewal has been set up to run monthly."
echo
echo "Your certificates are located at:"
echo "  - ./certs/live/$DOMAIN/fullchain.pem"
echo "  - ./certs/live/$DOMAIN/privkey.pem"
echo "  - ./certs/live/$DOMAIN/chain.pem" 