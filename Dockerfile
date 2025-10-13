# Stage 1: Build environment
FROM karalabe/xgo-1.13 AS builder

ENV ALSA_LIB_VERSION=1.2.4

RUN apt-get update && apt-get install -y \
    curl

RUN mkdir /alsa && \
    curl "ftp://ftp.alsa-project.org/pub/lib/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2" -o /alsa/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2

RUN cd /alsa && \
    tar -xvf alsa-lib-${ALSA_LIB_VERSION}.tar.bz2 && \
    cd alsa-lib-${ALSA_LIB_VERSION} && \
    CC=arm-linux-gnueabihf-gcc-5 ./configure --host=arm-linux && \
    make && \
    make install

# Copy source code
WORKDIR /go/src/github.com/hifi-baby/go-server
COPY . .

# Build the application for linux/arm with rpi tag
RUN CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CC=arm-linux-gnueabihf-gcc-5 go build -tags rpi -o /hifi-baby

# Stage 2: Runtime environment
FROM debian:buster-slim

# Create a non-root user
RUN useradd -ms /bin/bash hifi-baby

# Create app directory
RUN mkdir -p /home/hifi-baby/tracks /home/hifi-baby/dist
WORKDIR /home/hifi-baby

# Copy the binary from the builder stage
COPY --from=builder /hifi-baby .

# Copy other necessary files
COPY settings.json .
# Assuming ui/dist exists as it is part of the deployment script
COPY ui/dist ./dist

# Set ownership
RUN chown -R hifi-baby:hifi-baby /home/hifi-baby

USER hifi-baby

# Expose port
EXPOSE 3000

# Set environment variables
ENV SERVER_URL=0.0.0.0:3000
ENV SETTINGS_PATH=/home/hifi-baby/settings.json
ENV STORAGE_PATH=/home/hifi-baby/tracks
ENV SERVER_UI_PATH=/home/hifi-baby/dist
ENV DATABASE_PATH=/home/hifi-baby/hifi-baby.db
ENV DATABASE_TIMEOUT=10s

# Run the application
CMD ["./hifi-baby"]