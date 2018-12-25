FROM ilius/debian-go-py3:latest

ARG host

RUN if [ -z "$host" ] ; then echo "'host' argument not provided" ; exit 1 ; fi

ENV GOPATH /opt/go
ENV SRC $GOPATH/src/github.com/ilius/starcal-server

WORKDIR $SRC

COPY . $SRC

RUN mv -f $SRC/settings/hosts/${host}.py.docker $SRC/settings/hosts/${host}.py

RUN STARCAL_HOST=$host ./build.sh --no-remove
