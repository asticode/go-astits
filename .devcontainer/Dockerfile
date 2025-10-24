FROM alpine:latest

RUN apk add --update

##
# git
##

RUN apk add git openssh-client

##
# go
##

ADD https://dl.google.com/go/go1.25.3.linux-amd64.tar.gz /tmp/go.tar.gz
RUN tar -C /opt -xzf /tmp/go.tar.gz
ENV PATH="$PATH:/opt/go/bin"