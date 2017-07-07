#!/bin/bash
myDir=`dirname "$0"`
myDir=`realpath "$myDir"`
scalDir=`dirname "$myDir"`
srcDir=`dirname "$scalDir"`
rootDir=`dirname "$srcDir"`

if [ ! -f $srcDir/scal/settings/settings.go ] ; then
	STARCAL_HOST=localhost $rootDir/settings/build.py --no-build || exit $?
fi

GOPATH=$rootDir go build gen.go && ./gen "$@"
