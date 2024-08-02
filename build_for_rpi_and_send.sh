#!/bin/bash
# required
# sudo apt install libasound2-dev
# sudo apt-get install gcc-arm-linux-gnueabihf
#
# CC=arm-linux-gnueabihf-gcc ./configure --host=arm-linux && \
# make
# rm -rf ~/.cache/go-build
#
# LD_LIBRARY_PATH=/home/lmartin/Documents/Perso/hifi-baby/cross-compile/cross-pi-gcc-10.3.0-1/lib:$LD_LIBRARY_PATH \
# PATH=/home/lmartin/Documents/Perso/hifi-baby/cross-compile/cross-pi-gcc-10.3.0-1/bin:$PATH \
# CGO_CFLAGS="-L/home/lmartin/Documents/Perso/hifi-baby/ssh/mnt/usr/lib" \


PATH=/usr/arm-linux-gnueabihf/bin:$PATH \
CGO_LDFLAGS="-L/home/lmartin/Documents/Perso/hifi-baby/cross-compile/alsa-lib-1.2.7.2/src/.libs -lasound" \
CGO_CPPFLAGS="-I/home/lmartin/Documents/Perso/hifi-baby/cross-compile/alsa-lib-1.2.7.2/include" \
env CGO_ENABLED=1 \
CC=arm-linux-gnueabihf-gcc \
GOOS=linux \
GOARCH=arm \
GOARM=7 \
go build -o hifi-baby

# LD_LIBRARY_PATH=/home/lmartin/Documents/Perso/hifi-baby/cross-compile/cross-pi-gcc-10.3.0-1/lib:/home/lmartin/Documents/Perso/hifi-baby/ssh/mnt/usr/lib:$LD_LIBRARY_PATH \
# PATH=/usr/arm-linux-gnueabihf/bin:$PATH \
# CGO_LDFLAGS="-L/home/lmartin/Documents/Perso/hifi-baby/cross-compile/alsa-lib-1.2.7.2/src/.libs  -L/home/lmartin/Documents/Perso/hifi-baby/cross-compile/cross-pi-gcc-10.3.0-1/lib -lasound" \
# CGO_CPPFLAGS="-I/home/lmartin/Documents/Perso/hifi-baby/cross-compile/alsa-lib-1.2.7.2/include:/home/lmartin/Documents/Perso/hifi-baby/cross-compile/cross-pi-gcc-10.3.0-1/include" \
# env CGO_ENABLED=1 \
# CC=arm-linux-gnueabihf-gcc \
# GOOS=linux \
# GOARCH=arm \
# GOARM=7 \
# go build -o hifi-baby

scp -r hifi-baby stored_config.json ui/dist hifi-baby.service hifi-baby.default hifi-baby@192.168.1.76:/home/hifi-baby
