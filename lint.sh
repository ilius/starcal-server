#!/bin/bash
GOPATH=$PWD golint scal/... | less
