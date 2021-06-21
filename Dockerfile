FROM golang:1.16

RUN apt-get update -yy && \
    apt-get install -yy \
        autoconf \
        libtool \
        portaudio19-dev

COPY . /go/src/github.com/albanseurat/goplay2/
WORKDIR /go/src/github.com/albanseurat/goplay2/
RUN GOOS=linux make

FROM debian:jessie
RUN apt-get update -yy && \
    apt-get install -yy libportaudio2
WORKDIR /opt
COPY --from=0 /go/src/github.com/albanseurat/goplay2/goplay2 .
CMD ["/opt/goplay2"]


