#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	echo 'For example: export STARCAL_HOST=localhost'
	exit 1
fi

# takes ~ 0.2 seconds if submodules are already initialized / cloned
git submodule update --init

#GOPATH=$PWD go install github.com/globalsign/mgo
#GOPATH=$PWD go install golang.org/x/crypto/bcrypt
#GOPATH=$PWD go install golang.org/x/net/context

./settings/build.py "$@"
exit $?
