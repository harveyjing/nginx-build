FROM debian:bullseye-slim

# Set nginx version
ENV NGINX_VERSION=1.26.3
ENV BORINGSSL_CHECKSUM=c59bf8bf189dcbde868e04efcd53b705ed155231
ENV BORINGSSL_LOCAL="/build"

# Build-time metadata
ARG BUILD_DATE
ARG VCS_REF

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    libpcre3-dev \
    zlib1g-dev \
    wget \
    tar \
    gcc \
    make \
    libc-dev \
    git \
    cmake \
    golang \
    ninja-build \
    && rm -rf /var/lib/apt/lists/*

# Create directories
WORKDIR ${BORINGSSL_LOCAL}
RUN git clone --depth=1 --shallow-submodules --branch main https://boringssl.googlesource.com/boringssl \
  && cd boringssl \
  && mkdir build \
  && cd build \
  && cmake -GNinja -DBUILD_SHARED_LIBS=1 .. \
  && ninja

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
RUN ./configure \
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
    --with-ld-opt="-L$BORINGSSL_LOCAL/boringssl/build/ssl -L$BORINGSSL_LOCAL/boringssl/build/crypto" \
    --build="docker-nginx-http3-$VCS_REF-$BUILD_DATE" \
    && touch $BORINGSSL_LOCAL/boringssl/.openssl/include/openssl/ssl.h \
    && make -j$(nproc) \
    && make install

# Create a directory for the compiled binary
RUN mkdir -p /output

# Copy the compiled nginx binary to /output
RUN cp /usr/sbin/nginx /output/

# Set the output directory as the working directory
WORKDIR /output

# Install the 'file' command
RUN apt-get update && apt-get install -y file && rm -rf /var/lib/apt/lists/*

# Define the entrypoint to output information about the binary
CMD ["sh", "-c", "file nginx && ./nginx -V && echo 'Nginx binary is available at /output/nginx'"] 