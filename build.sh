#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	echo 'For example: export STARCAL_HOST=localhost'
	exit 1
fi

GOPATH=$PWD go install gopkg.in/mgo.v2
#GOPATH=$PWD go install golang.org/x/crypto/bcrypt
#GOPATH=$PWD go install golang.org/x/net/context

./settings/build.py "$@"
exit $?
