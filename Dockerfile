FROM karalabe/xgo-1.13

ENV ALSA_LIB_VERSION=1.2.4

RUN apt-get update && apt-get install -y \
    curl

RUN mkdir /alsa && \
    curl "ftp://ftp.alsa-project.org/pub/lib/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2" -o /alsa/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2

# https://www.programering.com/a/MTN0UDMwATk.html
# https://stackoverflow.com/questions/36195926/alsa-util-1-1-0-arm-cross-compile-issue
RUN cd /alsa && \
    tar -xvf alsa-lib-${ALSA_LIB_VERSION}.tar.bz2 && \
    cd alsa-lib-${ALSA_LIB_VERSION} && \
    CC=arm-linux-gnueabihf-gcc-5 ./configure --host=arm-linux && \
    make && \
    make install
