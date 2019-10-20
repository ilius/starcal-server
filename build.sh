#!/bin/bash

if [ -z "$NO_TOUCH_SUBMODULES" ] ; then
	# takes ~ 0.2 seconds if submodules are already initialized / cloned
	git submodule update --init
fi

#GOPATH=$PWD go install github.com/globalsign/mgo
#GOPATH=$PWD go install golang.org/x/crypto/bcrypt
#GOPATH=$PWD go install golang.org/x/net/context

./settings/build.py "$@"
exit $?
