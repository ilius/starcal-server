#!/bin/bash
find ./pkg/scal* -name '*.go' -exec gofmt -s -w '{}' \;
