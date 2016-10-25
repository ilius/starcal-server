#!/bin/bash
GOPATH=$PWD go install gopkg.in/mgo.v2-unstable
GOPATH=$PWD go install github.com/gorilla/context
GOPATH=$PWD go install github.com/gorilla/mux
GOPATH=$PWD go install golang.org/x/crypto/bcrypt
GOPATH=$PWD go install golang.org/x/net/context
GOPATH=$PWD go install scal-lib/go-http-auth
GOPATH=$PWD go build server-event.go
#GOPATH=$PWD go install scal
