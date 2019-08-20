#!/bin/bash

STARCAL_HOST=localhost ./settings/build.py --no-remove || exit 1
GOPATH=$PWD go test -v "$@"
rm src/scal/settings/settings.go
