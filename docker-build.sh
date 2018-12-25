#!/bin/bash
set -e

IMAGE="starcal-server-${STARCAL_HOST,,}"
TAG="latest"

# because "settings/hosts/$STARCAL_HOST.py" might be a symbolic link
cp -L "settings/hosts/$STARCAL_HOST.py" "settings/hosts/$STARCAL_HOST.py.docker"

docker build \
	-f Dockerfile \
	-t "$IMAGE:$TAG" \
	--build-arg "host=$STARCAL_HOST" \
	. || true

rm "settings/hosts/$STARCAL_HOST.py.docker"


