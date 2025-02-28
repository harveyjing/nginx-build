# Build Stage
FROM alpine:latest AS builder

# Set nginx version
ENV NGINX_VERSION=1.26.3
ENV BORINGSSL_LOCAL="/build"

# Build-time metadata
ARG BUILD_DATE
ARG VCS_REF=master

# Set BUILD_DATE if not provided at build time
RUN if [ -z "$BUILD_DATE" ]; then \
      export BUILD_DATE=$(date +%Y%m%d); \
      echo "BUILD_DATE not specified, using today's date: $BUILD_DATE"; \
    fi \
    && echo "BUILD_DATE=$BUILD_DATE" >> /build_info

# Install build dependencies
RUN apk update && apk add --no-cache \
    build-base \
    pcre-dev \
    zlib-dev \
    wget \
    tar \
    gcc \
    make \
    libc-dev \
    git \
    cmake \
    go \
    linux-headers \
    file

# Create directories
WORKDIR ${BORINGSSL_LOCAL}
RUN git clone --depth=1 --shallow-submodules --branch main https://boringssl.googlesource.com/boringssl \
  && cd boringssl \
  && mkdir build \
  && cd build \
  && cmake -DBUILD_SHARED_LIBS=1 .. \
  && make

# Prepare BoringSSL for Nginx compilation
RUN mkdir -p "$BORINGSSL_LOCAL/boringssl/.openssl/lib" \
&& cd "$BORINGSSL_LOCAL/boringssl/.openssl" \
&& ln -s ../include include \
&& cd "$BORINGSSL_LOCAL/boringssl" \
&& cp "build/crypto/libcrypto.so" ".openssl/lib" \
&& cp "build/ssl/libssl.so" ".openssl/lib" \
&& cp ".openssl/lib/libssl.so" /usr/lib/ \
&& cp ".openssl/lib/libcrypto.so" /usr/lib/

# Download and extract nginx source
WORKDIR /build
RUN wget http://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz \
    && tar -zxf nginx-${NGINX_VERSION}.tar.gz \
    && rm nginx-${NGINX_VERSION}.tar.gz

# Configure and build nginx with HTTP/3 support
WORKDIR /build/nginx-${NGINX_VERSION}
RUN source /build_info \
    && ./configure \
    --prefix=/usr/share/nginx \
    --sbin-path=/usr/sbin/nginx \
    --modules-path=/usr/lib/nginx/modules \
    --conf-path=/etc/nginx/nginx.conf \
    --error-log-path=/var/log/nginx/error.log \
    --http-log-path=/var/log/nginx/access.log \
    --pid-path=/var/run/nginx.pid \
    --lock-path=/var/run/nginx.lock \
    --http-client-body-temp-path=/var/cache/nginx/client_temp \
    --http-proxy-temp-path=/var/cache/nginx/proxy_temp \
    --http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp \
    --http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp \
    --http-scgi-temp-path=/var/cache/nginx/scgi_temp \
    --with-compat \
    --with-file-aio \
    --with-threads \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_dav_module \
    --with-http_flv_module \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_mp4_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-http_v3_module \
    --with-mail \
    --with-mail_ssl_module \
    --with-stream \
    --with-stream_realip_module \
    --with-stream_ssl_module \
    --with-stream_ssl_preread_module \
    --with-cc-opt="-I$BORINGSSL_LOCAL/boringssl/include -Wno-error" \
    --with-ld-opt="-L$BORINGSSL_LOCAL/boringssl/build/ssl -L$BORINGSSL_LOCAL/boringssl/build/crypto -Wl,-rpath,/usr/lib" \
    --build="docker-nginx-http3-$VCS_REF-$BUILD_DATE" \
    && touch $BORINGSSL_LOCAL/boringssl/.openssl/include/openssl/ssl.h \
    && make -j$(nproc) \
    && make install

# Create directories for configuration
RUN mkdir -p /etc/nginx/conf.d

# Copy configuration files
COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf

# Download mime.types file 
RUN wget -O /etc/nginx/mime.types https://raw.githubusercontent.com/nginx/nginx/master/conf/mime.types

# Create a simple HTML file for testing
RUN mkdir -p /usr/share/nginx/html \
    && echo "<html><body><h1>Nginx with HTTP/3 Support</h1><p>Powered by BoringSSL (Dynamic Build)</p></body></html>" > /usr/share/nginx/html/index.html

# Generate self-signed SSL certificates
RUN mkdir -p /etc/ssl/private \
    && apk add --no-cache openssl \
    && openssl req -x509 -newkey rsa:4096 -nodes \
       -keyout /etc/ssl/private/localhost.key \
       -out /etc/ssl/localhost.pem \
       -days 365 -sha256 -subj '/CN=localhost'

# Output nginx information for debugging
RUN file /usr/sbin/nginx && /usr/sbin/nginx -V

# Final Stage - Minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    pcre \
    zlib \
    ca-certificates \
    tzdata \
    libstdc++ \
    libgcc \
    && mkdir -p /var/cache/nginx \
    && mkdir -p /var/log/nginx \
    && mkdir -p /usr/share/nginx/html \
    && mkdir -p /etc/ssl/private

# Copy Nginx and config from builder stage
COPY --from=builder /usr/sbin/nginx /usr/sbin/nginx
COPY --from=builder /etc/nginx /etc/nginx
COPY --from=builder /usr/share/nginx/html /usr/share/nginx/html
COPY --from=builder /usr/lib/libssl.so /usr/lib/libssl.so
COPY --from=builder /usr/lib/libcrypto.so /usr/lib/libcrypto.so
COPY --from=builder /etc/ssl/localhost.pem /etc/ssl/localhost.pem
COPY --from=builder /etc/ssl/private/localhost.key /etc/ssl/private/localhost.key

# Forward request and error logs to docker log collector
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log

# Create nginx user and group
RUN addgroup -S nginx \
    && adduser -D -S -h /var/cache/nginx -s /sbin/nologin -G nginx nginx

# Ensure nginx user can write to required directories
RUN chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && mkdir -p /var/run \
    && chown -R nginx:nginx /var/run

# Expose HTTP, HTTPS, and HTTP/3 (QUIC) ports
EXPOSE 80 443/tcp 443/udp

# Set Nginx as the entrypoint
ENTRYPOINT ["nginx", "-g", "daemon off;"] 