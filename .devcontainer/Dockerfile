FROM alpine:latest

RUN apk add --update

##
# git
##

RUN apk add git openssh-client

##
# go
##

ARG GO_VERSION

ADD https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz /tmp/go.tar.gz
RUN tar -C /opt -xzf /tmp/go.tar.gz
ENV PATH="$PATH:/opt/go/bin"
