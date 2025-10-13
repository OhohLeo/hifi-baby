#!/bin/bash

set -e

# Default platform
PLATFORM="linux"

if [ -n "$1" ]; then
    PLATFORM=$1
fi

echo "Building for platform: $PLATFORM"

# Common build flags
BUILD_FLAGS="-o hifi-baby"

# Platform-specific settings
case "$PLATFORM" in
    "linux")
        export GOOS=linux
        export GOARCH=amd64
        ;;
    "rpi")
        # Raspberry Pi 2/3/4 (32-bit)
        export CGO_LDFLAGS="-L/home/lmartin/git/hifi-baby/cross-compile/alsa-lib-1.2.9/src/.libs -lasound"
        export CGO_CPPFLAGS="-I/home/lmartin/git/hifi-baby/cross-compile/alsa-lib-1.2.9/include"
        export CGO_ENABLED=1
        export CC=arm-linux-gnueabihf-gcc
        export GOOS=linux
        export GOARCH=arm
        export GOARM=7
        BUILD_FLAGS="$BUILD_FLAGS -tags rpi"
        ;;
    "rpi0")
        # Raspberry Pi Zero/1
        export CGO_LDFLAGS="-L/home/lmartin/git/hifi-baby/cross-compile/alsa-lib-1.2.9/src/.libs -lasound"
        export CGO_CPPFLAGS="-I/home/lmartin/git/hifi-baby/cross-compile/alsa-lib-1.2.9/include"
        export CGO_ENABLED=1
        export CC=arm-linux-gnueabihf-gcc
        export GOOS=linux
        export GOARCH=arm
        export GOARM=6
        BUILD_FLAGS="$BUILD_FLAGS -tags rpi"
        ;;
    "docker")
        echo "Building docker image"
        docker build -t hifi-baby .
        echo "Build complete."
        exit 0
        ;;
    *)
        echo "Unknown platform: $PLATFORM"
        echo "Supported platforms: linux, rpi, rpi0, docker"
        exit 1
        ;;
esac

echo "Building with flags: $BUILD_FLAGS"
go build $BUILD_FLAGS
echo "Build complete."
