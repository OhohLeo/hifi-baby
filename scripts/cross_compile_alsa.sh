#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <platform>"
    echo "Supported platforms: rpi, rpi0"
    exit 1
fi

PLATFORM=$1
ALSA_DIR="cross-compile/alsa-lib-1.2.9"

if [ ! -d "$ALSA_DIR" ]; then
    echo "ALSA directory not found at $ALSA_DIR"
    exit 1
fi

cd "$ALSA_DIR"

echo "Cleaning ALSA library..."
make clean && make distclean

echo "Configuring ALSA library for $PLATFORM..."

case "$PLATFORM" in
    "rpi")
        CC=arm-linux-gnueabihf-gcc ./configure --host=arm-linux
        ;;
    "rpi0")
        CC=aarch64-linux-gnu-gcc ./configure --host=aarch64-linux-gnu
        ;;
    *)
        echo "Unknown platform: $PLATFORM"
        echo "Supported platforms: rpi, rpi0"
        exit 1
        ;;
esac

echo "Building ALSA library..."
make

echo "ALSA library build complete."
