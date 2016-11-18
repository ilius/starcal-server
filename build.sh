#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
    echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
    echo 'For example: export STARCAL_HOST=localhost'
    exit 1
fi

GOPATH=$PWD go fmt scal/...

GOPATH=$PWD go install gopkg.in/mgo.v2
GOPATH=$PWD go install github.com/gorilla/context
GOPATH=$PWD go install github.com/gorilla/mux
GOPATH=$PWD go install golang.org/x/crypto/bcrypt
GOPATH=$PWD go install golang.org/x/net/context
GOPATH=$PWD go install scal-lib/go-http-auth

./settings/build.py

