# Single source stage
FROM    --platform=$BUILDPLATFORM golang:1.16-buster   as fetcher

# Bring the source in
COPY    . "$GOPATH"/src/github.com/albanseurat/goplay2
WORKDIR "$GOPATH"/src/github.com/albanseurat/goplay2

# Patch source to be able to link dynamically as debian packages do not provide static version of libfdk
RUN      sed -Ei 's/-l:libfdk-aac.a/-lfdk-aac/' ./codec/aac.go

# Actual cross-compiling image, per target platform
FROM    --platform=$BUILDPLATFORM fetcher       as builder

ARG     TARGETOS
ARG     TARGETARCH
ARG     TARGETVARIANT

ENV     GOOS=$TARGETOS
ENV     GOARCH=$TARGETARCH
ENV     CGO_ENABLED=1

RUN     sed -Ei 's/main$/main non-free/g' /etc/apt/sources.list; \
        DEB_TARGET_ARCH="$(echo "$TARGETARCH$TARGETVARIANT" | sed -e "s/^armv6$/armel/" -e "s/^armv7$/armhf/" -e "s/^ppc64le$/ppc64el/" -e "s/^386$/i386/")"; \
        dpkg --add-architecture "$DEB_TARGET_ARCH"; \
        apt-get update; \
        apt-get install -qq --no-install-recommends \
            autoconf \
            libtool \
            crossbuild-essential-"$DEB_TARGET_ARCH" \
            libfdk-aac1:"$DEB_TARGET_ARCH" \
            libfdk-aac-dev:"$DEB_TARGET_ARCH" \
            portaudio19-dev:"$DEB_TARGET_ARCH"

RUN     eval "$(dpkg-architecture -A "$(echo "$TARGETARCH$TARGETVARIANT" | sed -e "s/^armv6$/armel/" -e "s/^armv7$/armhf/" -e "s/^ppc64le$/ppc64el/" -e "s/^386$/i386/")")"; \
        export CC="${DEB_TARGET_GNU_TYPE}-gcc"; \
        export CXX="${DEB_TARGET_GNU_TYPE}-g++"; \
        export GOARM="$(printf "%s" "$TARGETVARIANT" | tr -d v)"; \
        go build -trimpath -buildmode=pie -tags "netgo osusergo cgo" -ldflags "-s -w" -o /dist/goplay2 .

# Runtime image
FROM    debian:buster

RUN     sed -Ei 's/main$/main non-free/g' /etc/apt/sources.list; \
        apt-get update; \
        apt-get install -qq --no-install-recommends \
            libportaudio2 \
            pulseaudio-utils \
            libfdk-aac1; \
        apt-get -qq autoremove      && \
        apt-get -qq clean           && \
        rm -rf /var/lib/apt/lists/* && \
        rm -rf /tmp/*               && \
        rm -rf /var/tmp/*

WORKDIR /opt

RUN     adduser --system --no-create-home --home /nonexistent --gecos "in dockerfile user" --uid 1000 goplay
RUN     chown 1000 /opt

COPY    --from=builder --chown=1000:root /dist/goplay2 .
RUN     setcap 'cap_net_bind_service+ep' ./goplay2

USER    goplay

CMD     ["/opt/goplay2"]
