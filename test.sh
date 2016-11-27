#!/bin/bash

STARCAL_HOST=localhost ./settings/build.py --no-remove
GOPATH=$PWD go test -v "$@"
rm src/scal/settings/settings.go
