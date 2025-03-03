#!/bin/bash
set -e

# Replace these with your domain and email
DOMAIN="wjing.xyz"
PRIMARY_DOMAIN="www.wjing.xyz"
EMAIL="wjing@wjing.dev"

CONTAINER_NAME="nginx-https-server"
IMAGE_NAME="nginx-http3"

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