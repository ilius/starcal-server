#!/bin/bash
find ./src/scal* -name '*.go' -exec gofmt -s -w '{}' \;
