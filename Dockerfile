FROM golang:1.18.1

ENV GOPROXY=https://goproxy.io,direct

WORKDIR /build

ADD . /build

ADD ./dockerdata/build/chrome/install /usr/local

RUN ls /build

